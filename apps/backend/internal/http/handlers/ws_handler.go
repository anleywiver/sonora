package handlers

import (
	"context"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appplayback "sonora.dev/go-core/application/playback"
	"sonora.dev/go-core/infrastructure/wstoken"

	"sonora.dev/backend/internal/http/middleware"
	"sonora.dev/backend/internal/http/response"
	"sonora.dev/backend/internal/ws"
)

const localsWSUserID = "ws_user_id"

type WSHandler struct {
	tokens   *wstoken.Issuer
	hub      *ws.Hub
	playback *appplayback.Service
}

func NewWSHandler(tokens *wstoken.Issuer, hub *ws.Hub, playback *appplayback.Service) *WSHandler {
	return &WSHandler{tokens: tokens, hub: hub, playback: playback}
}

// IssueToken hands out a 60-second single-use token for the WS handshake
// (ADR 0001 — the handshake can't send a custom Authorization header).
func (h *WSHandler) IssueToken(c *fiber.Ctx) error {
	token, err := h.tokens.Issue(c.Context(), middleware.UserID(c))
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to issue ws token")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{"token": token})
}

// UpgradeGate runs as normal HTTP middleware before the websocket.New
// handler — it consumes the token (single use) and stashes the resolved
// user ID for Handle to pick up.
func (h *WSHandler) UpgradeGate(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}
	userID, err := h.tokens.Consume(c.Context(), c.Query("token"))
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthenticated", "invalid or expired ws token")
	}
	c.Locals(localsWSUserID, userID)
	return c.Next()
}

func (h *WSHandler) Handle(conn *websocket.Conn) {
	userID, ok := conn.Locals(localsWSUserID).(uuid.UUID)
	if !ok {
		_ = conn.Close()
		return
	}

	h.hub.Register(userID, conn)
	defer h.hub.Unregister(userID, conn)

	// Push current state immediately so a device joining mid-session
	// doesn't have to wait for the next change elsewhere to sync up.
	if state, err := h.playback.GetState(context.Background(), userID); err == nil {
		_ = conn.WriteJSON(fiber.Map{"type": "player:state", "data": stateJSON(state)})
	}

	// Sprint 7 is broadcast-only — player:command (remote control from a
	// non-active device) is Sprint 8's Active Device pattern. For now we
	// just keep reading to detect disconnects.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
