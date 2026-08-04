// Package ws holds the in-process connection registry for broadcasting
// playback_state to every device a user has connected, and for relaying
// player:command from a remote-controller device to the Active Device. In-
// memory is correct for the target deployment (1 VPS, single api process —
// see CLAUDE.md "Batasan Deployment"); a multi-instance deployment would
// need a shared pub/sub (e.g. Redis) instead, which isn't in scope.
package ws

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

type connEntry struct {
	conn     *websocket.Conn
	deviceID uuid.UUID
}

type Hub struct {
	mu    sync.Mutex
	conns map[uuid.UUID]map[*websocket.Conn]uuid.UUID // userID -> conn -> deviceID
}

func NewHub() *Hub {
	return &Hub{conns: make(map[uuid.UUID]map[*websocket.Conn]uuid.UUID)}
}

func (h *Hub) Register(userID, deviceID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[userID] == nil {
		h.conns[userID] = make(map[*websocket.Conn]uuid.UUID)
	}
	h.conns[userID][conn] = deviceID
}

func (h *Hub) Unregister(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.conns[userID], conn)
	if len(h.conns[userID]) == 0 {
		delete(h.conns, userID)
	}
}

// Broadcast sends message as JSON to every connection registered for
// userID. A write failure on one connection (e.g. it just dropped) is
// swallowed — its own read loop will unregister it shortly.
func (h *Hub) Broadcast(userID uuid.UUID, message any) {
	for _, c := range h.connsFor(userID) {
		_ = c.conn.WriteJSON(message)
	}
}

// SendToDevice relays message only to userID's connection for a specific
// device (used for player:command → the Active Device). Returns false if
// that device has no open connection right now.
func (h *Hub) SendToDevice(userID, deviceID uuid.UUID, message any) bool {
	for _, c := range h.connsFor(userID) {
		if c.deviceID == deviceID {
			_ = c.conn.WriteJSON(message)
			return true
		}
	}
	return false
}

func (h *Hub) connsFor(userID uuid.UUID) []connEntry {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]connEntry, 0, len(h.conns[userID]))
	for c, deviceID := range h.conns[userID] {
		out = append(out, connEntry{conn: c, deviceID: deviceID})
	}
	return out
}
