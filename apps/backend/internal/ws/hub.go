// Package ws holds the in-process connection registry for broadcasting
// playback_state to every device a user has connected. In-memory is
// correct for the target deployment (1 VPS, single api process — see
// CLAUDE.md "Batasan Deployment"); a multi-instance deployment would need
// a shared pub/sub (e.g. Redis) instead, which isn't in scope.
package ws

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
)

type Hub struct {
	mu    sync.Mutex
	conns map[uuid.UUID]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{conns: make(map[uuid.UUID]map[*websocket.Conn]struct{})}
}

func (h *Hub) Register(userID uuid.UUID, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.conns[userID] == nil {
		h.conns[userID] = make(map[*websocket.Conn]struct{})
	}
	h.conns[userID][conn] = struct{}{}
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
	h.mu.Lock()
	conns := make([]*websocket.Conn, 0, len(h.conns[userID]))
	for c := range h.conns[userID] {
		conns = append(conns, c)
	}
	h.mu.Unlock()

	for _, c := range conns {
		_ = c.WriteJSON(message)
	}
}
