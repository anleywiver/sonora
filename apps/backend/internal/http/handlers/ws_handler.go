package handlers

import (
	"context"
	"encoding/json"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	appplayback "sonora.dev/go-core/application/playback"
	"sonora.dev/go-core/infrastructure/wstoken"

	"sonora.dev/backend/internal/http/middleware"
	"sonora.dev/backend/internal/http/response"
	"sonora.dev/backend/internal/ws"
)

const (
	localsWSUserID   = "ws_user_id"
	localsWSDeviceID = "ws_device_id"
)

type WSHandler struct {
	tokens   *wstoken.Issuer
	hub      *ws.Hub
	playback *appplayback.Service
}

func NewWSHandler(tokens *wstoken.Issuer, hub *ws.Hub, playback *appplayback.Service) *WSHandler {
	return &WSHandler{tokens: tokens, hub: hub, playback: playback}
}

type issueWSTokenRequest struct {
	DeviceID string `json:"device_id"`
}

// IssueToken hands out a 60-second single-use token for the WS handshake
// (ADR 0001 — the handshake can't send a custom Authorization header).
// The caller's own device_id is embedded so the hub can later target
// player:command relays at a specific device (the Active Device).
func (h *WSHandler) IssueToken(c *fiber.Ctx) error {
	var req issueWSTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid request body")
	}
	deviceID, err := uuid.Parse(req.DeviceID)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "validation_error", "invalid device_id")
	}

	token, err := h.tokens.Issue(c.Context(), middleware.UserID(c), deviceID)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "internal_error", "failed to issue ws token")
	}
	return response.OK(c, fiber.StatusOK, fiber.Map{"token": token})
}

// UpgradeGate runs as normal HTTP middleware before the websocket.New
// handler — it consumes the token (single use) and stashes the resolved
// user/device ID for Handle to pick up.
func (h *WSHandler) UpgradeGate(c *fiber.Ctx) error {
	if !websocket.IsWebSocketUpgrade(c) {
		return fiber.ErrUpgradeRequired
	}
	claims, err := h.tokens.Consume(c.Context(), c.Query("token"))
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "unauthenticated", "invalid or expired ws token")
	}
	c.Locals(localsWSUserID, claims.UserID)
	c.Locals(localsWSDeviceID, claims.DeviceID)
	return c.Next()
}

type inboundMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type commandData struct {
	Command string          `json:"command"`
	Payload json.RawMessage `json:"payload"`
}

func (h *WSHandler) Handle(conn *websocket.Conn) {
	userID, ok := conn.Locals(localsWSUserID).(uuid.UUID)
	if !ok {
		_ = conn.Close()
		return
	}
	deviceID, ok := conn.Locals(localsWSDeviceID).(uuid.UUID)
	if !ok {
		_ = conn.Close()
		return
	}

	h.hub.Register(userID, deviceID, conn)
	defer h.hub.Unregister(userID, conn)

	// Push current state immediately so a device joining mid-session
	// doesn't have to wait for the next change elsewhere to sync up.
	if state, err := h.playback.GetState(context.Background(), userID); err == nil {
		_ = conn.WriteJSON(fiber.Map{"type": "player:state", "data": stateJSON(state)})
	}

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}
		h.handleInbound(context.Background(), userID, deviceID, raw)
	}
}

// handleInbound relays player:command from a remote-controller device to
// whichever device is currently Active — that device's own client-side
// code executes the command locally (it has the real <audio> element) and
// reports the resulting state back via POST /player/state. This handler
// never touches playback_state itself; it's purely a relay.
func (h *WSHandler) handleInbound(ctx context.Context, userID, senderDeviceID uuid.UUID, raw []byte) {
	var msg inboundMessage
	if err := json.Unmarshal(raw, &msg); err != nil || msg.Type != "player:command" {
		return
	}

	state, err := h.playback.GetState(ctx, userID)
	if err != nil || state.ActiveDeviceID == nil {
		return // nothing is active to command
	}
	if *state.ActiveDeviceID == senderDeviceID {
		return // the active device commands itself directly, not via relay
	}

	var cmd commandData
	if err := json.Unmarshal(msg.Data, &cmd); err != nil {
		return
	}
	h.hub.SendToDevice(userID, *state.ActiveDeviceID, fiber.Map{
		"type": "player:command",
		"data": cmd,
	})
}
