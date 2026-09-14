package room

import (
	"sync"
)

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*Room),
	}
}

func (h *Hub) GetOrCreate(id string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()

	if r, exists := h.rooms[id]; exists {
		return r
	}

	r := New(id)
	h.rooms[id] = r
	return r
}

func (h *Hub) Get(id string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[id]
}
