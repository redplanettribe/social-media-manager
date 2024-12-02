package session

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
	UpdatedAt time.Time
}

func NewSession(userID string) *Session {
	return &Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
	}
}
