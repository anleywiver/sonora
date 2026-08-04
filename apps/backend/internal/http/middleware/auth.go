package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"sonora.dev/go-core/infrastructure/jwt"

	"sonora.dev/backend/internal/http/response"
)

const (
	localsUserID = "user_id"
	localsRole   = "role"
)

// RequireAuth parses the Bearer access token and stores user_id/role in
// fiber.Locals for downstream handlers. Streaming/WS endpoints use their
// own short-lived scoped tokens instead — this only guards normal API
// routes.
func RequireAuth(issuer *jwt.Issuer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		const prefix = "Bearer "
		if !strings.HasPrefix(header, prefix) {
			return response.Fail(c, fiber.StatusUnauthorized, "unauthenticated", "missing bearer token")
		}

		claims, err := issuer.Parse(strings.TrimPrefix(header, prefix))
		if err != nil {
			return response.Fail(c, fiber.StatusUnauthorized, "unauthenticated", "invalid or expired token")
		}

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return response.Fail(c, fiber.StatusUnauthorized, "unauthenticated", "invalid token subject")
		}

		c.Locals(localsUserID, userID)
		c.Locals(localsRole, claims.Role)
		return c.Next()
	}
}

// RequireRole must run after RequireAuth.
func RequireRole(role string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Locals(localsRole) != role {
			return response.Fail(c, fiber.StatusForbidden, "forbidden", "insufficient role")
		}
		return c.Next()
	}
}

func UserID(c *fiber.Ctx) uuid.UUID {
	id, _ := c.Locals(localsUserID).(uuid.UUID)
	return id
}

func UserRole(c *fiber.Ctx) string {
	role, _ := c.Locals(localsRole).(string)
	return role
}
