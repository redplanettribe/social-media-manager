package session

import (
	"context"
	"crypto/sha256"
	"fmt"
)

//go:generate mockery --name=Repository --case=underscore --inpackage
type Repository interface {
	// CreateSession creates a new session for the user with the given ID.
	CreateSession(ctx context.Context, session *Session) (string, error)
	// DeleteSessionsForUser deletes all sessions for the user with the given ID.
	DeleteSessionsForUser(ctx context.Context, userID string) error
	// DeleteSession deletes the session with the given ID.
	DeleteSession(ctx context.Context, sessionID string) error
	// GetSessionByID retrieves the session with the given ID.
	GetSessionByID(ctx context.Context, sessionID string) (*Session, error)
}

//go:generate mockery --name=Manager --case=underscore --inpackage
type Manager interface {
	// CreateSession creates a new session for the user with the given ID.
	CreateSession(ctx context.Context, userID, fingerprint string) (*Session, error)
	// ValidateSession checks if the session is valid.
	ValidateSession(ctx context.Context, sessionID, fingerprint string) (*Session, error)
	// InvalidateSession invalidates the session with the given ID.
	InvalidateSession(ctx context.Context, sessionID string) error
}

/*
We could generate a token, but at the moment we are just going to use a UUID. Maybe in combination with device fingerprinting.
*/
type manager struct {
	repo Repository
}

func NewManager(repo Repository) Manager {
	return &manager{
		repo: repo,
	}
}

func (m *manager) CreateSession(ctx context.Context, userID, fingerprint string) (*Session, error) {
	hashedFingerprint := hashFingerprint(fingerprint)
	err := m.repo.DeleteSessionsForUser(ctx, userID)
	if err != nil {
		return &Session{}, err
	}
	session := NewSession(userID, hashedFingerprint)
	_, err = m.repo.CreateSession(ctx, session)
	if err != nil {
		return &Session{}, err
	}
	return session, nil
}

func (m *manager) ValidateSession(ctx context.Context, sessionID, fingerprint string) (*Session, error) {
	session, err := m.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return &Session{}, err
	}
	err = validateFingerprint(session.DeviceFingerprint, fingerprint)
	if err != nil {
		return &Session{}, err
	}
	if !session.IsValid() {
		return &Session{}, ErrInvalidSession
	}
	return session, nil
}

func (m *manager) InvalidateSession(ctx context.Context, sessionID string) error {
	return m.repo.DeleteSession(ctx, sessionID)
}

func hashFingerprint(fingerprint string) string {
	hash := sha256.Sum256([]byte(fingerprint))
	return fmt.Sprintf("%x", hash)
}

func validateFingerprint(hashedFingerprint, fingerprint string) error {
	hash := sha256.Sum256([]byte(fingerprint))
	if hashedFingerprint != fmt.Sprintf("%x", hash) {
		return ErrInvalidFingerprint
	}
	return nil
}
