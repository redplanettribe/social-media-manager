package session

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidSession     = errors.New("invalid session")
	ErrInvalidFingerprint = errors.New("invalid fingerprint")
	ErrNoFingerprint      = errors.New("no fingerprint")
)

type contextKey string

const (
	ClientIPKey  contextKey = "clientIP"
	UserAgentKey contextKey = "userAgent"
	UserIDKey    contextKey = "userID"
)

type Session struct {
	ID                string
	UserID            string
	DeviceFingerprint string
	CreatedAt         time.Time
	ExpiresAt         time.Time
	UpdatedAt         time.Time
}

func NewSession(userID, deviceFingerprint string) *Session {
	return &Session{
		ID:                uuid.New().String(),
		UserID:            userID,
		DeviceFingerprint: deviceFingerprint,
		CreatedAt:         time.Now().UTC(),
		ExpiresAt:         time.Now().Add(time.Hour * 24 * 7).UTC(),
	}
}

func (s *Session) IsValid() bool {
	return time.Now().UTC().Before(s.ExpiresAt)
}
