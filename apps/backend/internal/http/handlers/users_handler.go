package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appusers "sonora.dev/go-core/application/users"

	"sonora.dev/backend/internal/http/response"
)

// UsersHandler backs the admin Manage Users page (Sprint 14 sisipan,
// ADR 0009).
type UsersHandler struct {
	service *appusers.Service
}

func NewUsersHandler(service *appusers.Service) *UsersHandler {
	return &UsersHandler{service: service}
}

func (h *UsersHandler) List(c *fiber.Ctx) error {
	list, err := h.service.List(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to list users")
	}
	out := make([]fiber.Map, 0, len(list))
	for _, u := range list {
		status := "active"
		if u.IsPending {
			status = "invited"
		}
		out = append(out, fiber.Map{
			"id": u.ID, "name": u.Name, "email": u.Email, "role": u.Role,
			"status": status, "created_at": u.CreatedAt,
		})
	}
	return response.OK(c, fiber.StatusOK, out)
}

type inviteUserRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (h *UsersHandler) Invite(c *fiber.Ctx) error {
	var req inviteUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	if req.Email == "" {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "email is required")
	}
	user, err := h.service.Invite(c.Context(), req.Email, req.Name)
	if err != nil {
		if errors.Is(err, appusers.ErrEmailTaken) {
			return response.Fail(c, fiber.StatusConflict, "conflict", "a user with this email already exists")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to invite user")
	}
	return response.OK(c, fiber.StatusCreated, fiber.Map{
		"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role, "status": "invited",
	})
}

type createUserRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// Create is the "Add User" form (Sprint 14 sisipan, ADR 0012) —
// credential-based Member, active immediately. Separate from Invite
// (which stays the Google-login path, ADR 0009).
func (h *UsersHandler) Create(c *fiber.Ctx) error {
	var req createUserRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	if req.Username == "" || req.Password == "" {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "username and password are required")
	}
	if len(req.Password) < 8 {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "password must be at least 8 characters")
	}
	user, err := h.service.CreateWithPassword(c.Context(), req.Username, req.Name, req.Password, req.Email)
	if err != nil {
		if errors.Is(err, appusers.ErrUsernameTaken) {
			return response.Fail(c, fiber.StatusConflict, "conflict", "a user with this username already exists")
		}
		if errors.Is(err, appusers.ErrEmailTaken) {
			return response.Fail(c, fiber.StatusConflict, "conflict", "a user with this email already exists")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to create user")
	}
	return response.OK(c, fiber.StatusCreated, fiber.Map{
		"id": user.ID, "name": user.Name, "email": user.Email, "role": user.Role, "status": "active",
	})
}

func (h *UsersHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid user id")
	}
	if err := h.service.RemoveAccess(c.Context(), id); err != nil {
		if errors.Is(err, appusers.ErrNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "not_found", "user not found")
		}
		if errors.Is(err, appusers.ErrCannotRemoveOwner) {
			return response.Fail(c, fiber.StatusForbidden, "forbidden", "cannot remove an Owner's access")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to remove user")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{})
}
