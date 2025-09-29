package relay

import (
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Hub struct {
	mu      sync.RWMutex
	sessions map[string]*SessionActor
	ttl      time.Duration
	idleTO   time.Duration
	ringSize int
	rateBPS  int
	redisURL string
	rdb      *redis.Client
}

func NewHub(ttl, idleTO time.Duration, ringSize, rateBPS int, redisURL string) *Hub {
	h := &Hub{
		sessions: make(map[string]*SessionActor),
		ttl:      ttl,
		idleTO:   idleTO,
		ringSize: ringSize,
		rateBPS:  rateBPS,
		redisURL: redisURL,
	}

	if redisURL != "" {
		// Parse Redis URL, assume redis://host:port
		parts := strings.Split(redisURL, "://")
		if len(parts) == 2 {
			addr := parts[1]
			h.rdb = redis.NewClient(&redis.Options{Addr: addr})
		}
	}

	return h
}

func (h *Hub) GetOrCreate(sid, tid string) *SessionActor {
	h.mu.Lock()
	defer h.mu.Unlock()

	if sa, ok := h.sessions[sid]; ok {
		return sa
	}

	sa := NewSessionActor(sid, tid, h.ttl, h.idleTO, h.ringSize, h.rateBPS, h.redisURL, h.rdb)

	h.sessions[sid] = sa
	IncActiveSessions(1)

	return sa
}

func (h *Hub) Close(sid, reason string) {
	h.mu.Lock()
	sa, ok := h.sessions[sid]
	delete(h.sessions, sid)
	h.mu.Unlock()

	if ok {
		sa.Close(reason)
		IncActiveSessions(-1)
	}
}

func (h *Hub) CloseAll() {
	h.mu.Lock()
	sessions := h.sessions
	h.sessions = make(map[string]*SessionActor)
	h.mu.Unlock()

	for _, sa := range sessions {
		sa.Close("shutdown")
	}
	IncActiveSessions(-len(sessions))
}
