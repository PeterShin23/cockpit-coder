package session

import (
	"context"
	"time"
)

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	Repo      string    `json:"repo"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Manager interface for session management
type Manager interface {
	Create(ctx context.Context, repo string, ttl time.Duration) (Session, error)
	Get(ctx context.Context, id string) (Session, bool)
	Touch(ctx context.Context, id string) error
	End(ctx context.Context, id string) error
}
