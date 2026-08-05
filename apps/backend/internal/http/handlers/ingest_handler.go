package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	appingest "sonora.dev/go-core/application/ingest"
	"sonora.dev/go-core/infrastructure/idempotency"

	"sonora.dev/backend/internal/http/middleware"
	"sonora.dev/backend/internal/http/response"
)

const idempotencyScopeIngestUpload = "ingest_upload"

type IngestHandler struct {
	service     *appingest.Service
	asynqClient *asynq.Client
	idempotency *idempotency.Store
	tmpDir      string
}

func NewIngestHandler(service *appingest.Service, asynqClient *asynq.Client, idem *idempotency.Store, tmpDir string) *IngestHandler {
	return &IngestHandler{service: service, asynqClient: asynqClient, idempotency: idem, tmpDir: tmpDir}
}

var filenameSanitizer = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// sanitizeFilename strips directory components and anything but a safe
// charset, since the result becomes part of a path on disk.
func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = filenameSanitizer.ReplaceAllString(name, "_")
	if name == "" || name == "." || name == ".." {
		name = "upload"
	}
	if len(name) > 100 {
		name = name[len(name)-100:]
	}
	return name
}

// Upload streams the multipart file straight to disk while hashing it
// (direct-to-backend streaming per ADR 0001, not a presigned URL), then
// hands off to the ingest service. A job with Status "pending" still needs
// its Asynq task enqueued; "completed" means checksum dedup already
// resolved it to an existing song.
func (h *IngestHandler) Upload(c *fiber.Ctx) error {
	userID := middleware.UserID(c)

	idempotencyKey := c.Get("Idempotency-Key")
	if idempotencyKey != "" {
		if jobID, found, err := h.idempotency.Lookup(c.Context(), idempotencyScopeIngestUpload, userID.String(), idempotencyKey); err == nil && found {
			return response.OK(c, fiber.StatusOK, fiber.Map{"id": jobID, "idempotent_replay": true})
		}
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "missing file field")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to read upload")
	}
	defer src.Close()

	if err := os.MkdirAll(h.tmpDir, 0o755); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to prepare upload storage")
	}

	safeName := sanitizeFilename(fileHeader.Filename)
	dst, err := os.CreateTemp(h.tmpDir, "ingest-*__"+safeName)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to prepare upload storage")
	}
	tempPath := dst.Name()

	hasher := sha256.New()
	if _, err := io.Copy(io.MultiWriter(dst, hasher), src); err != nil {
		dst.Close()
		_ = os.Remove(tempPath)
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to store upload")
	}
	dst.Close()
	checksum := hex.EncodeToString(hasher.Sum(nil))

	job, err := h.service.Accept(c.Context(), userID, "manual_upload", tempPath, checksum)
	if err != nil {
		_ = os.Remove(tempPath)
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to accept upload")
	}

	if job.Status == "pending" {
		if err := h.enqueue(job.ID); err != nil {
			return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to queue ingest job")
		}
	}

	if idempotencyKey != "" {
		_ = h.idempotency.Save(c.Context(), idempotencyScopeIngestUpload, userID.String(), idempotencyKey, job.ID.String())
	}

	return response.OK(c, fiber.StatusCreated, jobJSON(job))
}

func (h *IngestHandler) enqueue(jobID uuid.UUID) error {
	task, err := appingest.NewProcessTask(jobID)
	if err != nil {
		return err
	}
	_, err = h.asynqClient.Enqueue(task)
	return err
}

func (h *IngestHandler) List(c *fiber.Ctx) error {
	userID := middleware.UserID(c)
	status := c.Query("status")
	cursor := c.Query("cursor")
	limit := 20
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	jobs, nextCursor, hasMore, err := h.service.ListJobs(c.Context(), userID, status, cursor, int32(limit))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid cursor")
	}

	out := make([]fiber.Map, 0, len(jobs))
	for _, job := range jobs {
		out = append(out, jobJSON(job))
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":        out,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}

// AdminList is List without the per-user scope — the admin Job Queue page
// (Sprint 14, docs/screens-spec.md #20).
func (h *IngestHandler) AdminList(c *fiber.Ctx) error {
	status := c.Query("status")
	cursor := c.Query("cursor")
	limit := 20
	if raw := c.Query("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	jobs, nextCursor, hasMore, err := h.service.ListAllJobs(c.Context(), status, cursor, int32(limit))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid cursor")
	}

	out := make([]fiber.Map, 0, len(jobs))
	for _, job := range jobs {
		row := jobJSON(job)
		row["user_id"] = job.UserID
		out = append(out, row)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":        out,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}

// AdminRetry is Retry without the ownership check.
func (h *IngestHandler) AdminRetry(c *fiber.Ctx) error {
	jobID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid job id")
	}
	job, err := h.service.RetryJobAdmin(c.Context(), jobID)
	if err != nil {
		return h.jobError(c, err)
	}
	if err := h.enqueue(job.ID); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to queue retry")
	}
	return response.OK(c, fiber.StatusOK, jobJSON(job))
}

func (h *IngestHandler) Get(c *fiber.Ctx) error {
	jobID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid job id")
	}
	job, err := h.service.GetJob(c.Context(), jobID, middleware.UserID(c))
	if err != nil {
		return h.jobError(c, err)
	}
	return response.OK(c, fiber.StatusOK, jobJSON(job))
}

func (h *IngestHandler) Retry(c *fiber.Ctx) error {
	jobID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid job id")
	}
	job, err := h.service.RetryJob(c.Context(), jobID, middleware.UserID(c))
	if err != nil {
		return h.jobError(c, err)
	}
	if err := h.enqueue(job.ID); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to queue retry")
	}
	return response.OK(c, fiber.StatusOK, jobJSON(job))
}

func (h *IngestHandler) Delete(c *fiber.Ctx) error {
	jobID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid job id")
	}
	if err := h.service.DeleteJob(c.Context(), jobID, middleware.UserID(c)); err != nil {
		return h.jobError(c, err)
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}

func (h *IngestHandler) jobError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, appingest.ErrNotFound):
		return response.Fail(c, fiber.StatusNotFound, "not_found", "ingest job not found")
	case errors.Is(err, appingest.ErrNotRetryable):
		return response.Fail(c, fiber.StatusConflict, "conflict", "job is not in a retryable state")
	default:
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "ingest job operation failed")
	}
}

func jobJSON(j *appingest.Job) fiber.Map {
	return fiber.Map{
		"id":            j.ID,
		"source_type":   j.SourceType,
		"status":        j.Status,
		"song_id":       j.SongID,
		"error_message": j.ErrorMessage,
		"created_at":    j.CreatedAt,
		"updated_at":    j.UpdatedAt,
	}
}
