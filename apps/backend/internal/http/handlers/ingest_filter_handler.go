package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appingestfilter "sonora.dev/go-core/application/ingestfilter"

	"sonora.dev/backend/internal/http/response"
)

// IngestFilterHandler backs the admin Ingest Sources "Filter Rules" panel
// (Sprint 14 sisipan, ADR 0008) — genre/year rules for bandcamp/cloud_sync
// auto-ingest only, never manual_upload.
type IngestFilterHandler struct {
	service *appingestfilter.Service
}

func NewIngestFilterHandler(service *appingestfilter.Service) *IngestFilterHandler {
	return &IngestFilterHandler{service: service}
}

func (h *IngestFilterHandler) validSourceType(c *fiber.Ctx) (string, bool) {
	sourceType := c.Params("source_type")
	return sourceType, sourceType == "bandcamp" || sourceType == "cloud_sync"
}

func (h *IngestFilterHandler) List(c *fiber.Ctx) error {
	sourceType, ok := h.validSourceType(c)
	if !ok {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "source_type must be bandcamp or cloud_sync")
	}
	rules, err := h.service.ListRules(c.Context(), sourceType)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to list filter rules")
	}
	out := make([]fiber.Map, 0, len(rules))
	for _, r := range rules {
		out = append(out, fiber.Map{"id": r.ID, "source_type": r.SourceType, "rule_type": r.RuleType, "value": r.Value})
	}
	return response.OK(c, fiber.StatusOK, out)
}

type createFilterRuleRequest struct {
	RuleType string `json:"rule_type"`
	Value    string `json:"value"`
}

func (h *IngestFilterHandler) Create(c *fiber.Ctx) error {
	sourceType, ok := h.validSourceType(c)
	if !ok {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "source_type must be bandcamp or cloud_sync")
	}
	var req createFilterRuleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	if req.RuleType != "genre_allow" && req.RuleType != "year_min" && req.RuleType != "year_max" {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "rule_type must be genre_allow, year_min, or year_max")
	}
	if req.Value == "" {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "value is required")
	}
	rule, err := h.service.CreateRule(c.Context(), sourceType, req.RuleType, req.Value)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to create filter rule")
	}
	return response.OK(c, fiber.StatusCreated, fiber.Map{
		"id": rule.ID, "source_type": rule.SourceType, "rule_type": rule.RuleType, "value": rule.Value,
	})
}

func (h *IngestFilterHandler) Delete(c *fiber.Ctx) error {
	sourceType, ok := h.validSourceType(c)
	if !ok {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "source_type must be bandcamp or cloud_sync")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid rule id")
	}
	if err := h.service.DeleteRule(c.Context(), sourceType, id); err != nil {
		if errors.Is(err, appingestfilter.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "filter rule not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to delete filter rule")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}
