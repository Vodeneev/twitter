// Package realtime is a tiny in-memory pub/sub hub keyed by user id. It powers
// live notifications and direct messages over WebSocket. For a single-node
// deployment this is enough; a multi-node setup would swap it for Redis pub/sub.
package realtime

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type Client struct {
	UserID uuid.UUID
	Send   chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[uuid.UUID]map[*Client]struct{})}
}

func (h *Hub) Add(userID uuid.UUID) *Client {
	c := &Client{UserID: userID, Send: make(chan []byte, 16)}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*Client]struct{})
	}
	h.clients[userID][c] = struct{}{}
	return c
}

func (h *Hub) Remove(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set, ok := h.clients[c.UserID]; ok {
		if _, ok := set[c]; ok {
			delete(set, c)
			close(c.Send)
		}
		if len(set) == 0 {
			delete(h.clients, c.UserID)
		}
	}
}

// Publish delivers an event to every live connection of a user. Slow consumers
// are dropped rather than blocking the publisher.
func (h *Hub) Publish(userID uuid.UUID, ev Event) {
	payload, err := json.Marshal(ev)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients[userID] {
		select {
		case c.Send <- payload:
		default:
		}
	}
}
