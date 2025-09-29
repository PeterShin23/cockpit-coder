package events

import (
	"sync"
)

// Event represents a system event
type Event struct {
	Type   string         `json:"type"`
	Fields map[string]any `json:"fields,omitempty"`
}

// Bus interface for event publishing and subscription
type Bus interface {
	Subscribe(sessionID string) (<-chan Event, func())
	Publish(sessionID string, e Event)
}

// MemoryBus implements an in-memory event bus
type MemoryBus struct {
	topics map[string]map[chan Event]bool
	mu     sync.RWMutex
}

// NewMemoryBus creates a new in-memory event bus
func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		topics: make(map[string]map[chan Event]bool),
	}
}

// Subscribe creates a new subscription for a session
func (b *MemoryBus) Subscribe(sessionID string) (<-chan Event, func()) {
	ch := make(chan Event, 100) // Buffered channel to prevent blocking

	b.mu.Lock()
	if b.topics[sessionID] == nil {
		b.topics[sessionID] = make(map[chan Event]bool)
	}
	b.topics[sessionID][ch] = true
	b.mu.Unlock()

	// Return unsubscribe function
	unsubscribe := func() {
		b.mu.Lock()
		if subscribers, ok := b.topics[sessionID]; ok {
			delete(subscribers, ch)
			if len(subscribers) == 0 {
				delete(b.topics, sessionID)
			}
		}
		b.mu.Unlock()
		close(ch)
	}

	return ch, unsubscribe
}

// Publish sends an event to all subscribers of a session
func (b *MemoryBus) Publish(sessionID string, e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subscribers, ok := b.topics[sessionID]; ok {
		for ch := range subscribers {
			select {
			case ch <- e:
			default:
				// Channel is full, drop the event
			}
		}
	}
}
