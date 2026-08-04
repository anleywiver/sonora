package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"

	appbackup "sonora.dev/go-core/application/backup"

	"sonora.dev/backend/internal/http/response"
)

// BackupHandler lets the Owner trigger the Sprint 13 scheduled backup
// on demand (ADR 0007) — same "manual trigger alongside the schedule"
// pattern as Drive Manager's health-check and Ingest Sources' sync.
type BackupHandler struct {
	asynqClient *asynq.Client
}

func NewBackupHandler(asynqClient *asynq.Client) *BackupHandler {
	return &BackupHandler{asynqClient: asynqClient}
}

func (h *BackupHandler) Run(c *fiber.Ctx) error {
	if _, err := h.asynqClient.Enqueue(asynq.NewTask(appbackup.TaskTypeRunBackup, nil)); err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to queue backup job")
	}
	return response.OK(c, fiber.StatusAccepted, fiber.Map{})
}
