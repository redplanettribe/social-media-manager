package session

import "context"

//go:generate mockery --name=Repository --case=underscore --inpackage
type Repository interface {
	// CreateSession creates a new session for the user with the given ID.
	CreateSession(ctx context.Context, session *Session) (string, error)
	// DeleteSessionsForUser deletes all sessions for the user with the given ID.
	DeleteSessionsForUser(ctx context.Context, userID string) error
}

//go:generate mockery --name=Manager --case=underscore --inpackage
type Manager interface {
	// CreateSession creates a new session for the user with the given ID.
	CreateSession(ctx context.Context, userID string) (*Session, error)
	// ValidateSession checks if the session is valid.
	ValidateSession(ctx context.Context, sessionID string) (string, error)
}

type manager struct {
	repo Repository
}

func NewManager(repo Repository) Manager {
	return &manager{
		repo: repo,
	}
}

func (m *manager) CreateSession(ctx context.Context, userID string) (*Session, error) {
	// Invalidate all sessions for the user
	err := m.repo.DeleteSessionsForUser(ctx, userID)
	if err != nil {
		return &Session{}, err
	}
	// Create a new session
	session := NewSession(userID)
	_, err = m.repo.CreateSession(ctx, session)
	if err != nil {
		return &Session{}, err
	}
	return session, nil
}

func (m *manager) ValidateSession(ctx context.Context, sessionID string) (string, error) {
	// To be implemented
	return "", nil
}
