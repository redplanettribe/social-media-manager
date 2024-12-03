package session

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidSession = errors.New("invalid session")
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

func (s *Session) IsValid() bool {
	return time.Now().Before(s.ExpiresAt)
}
