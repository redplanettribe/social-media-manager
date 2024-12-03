package session

import "context"

//go:generate mockery --name=Repository --case=underscore --inpackage
type Repository interface {
	// CreateSession creates a new session for the user with the given ID.
	CreateSession(ctx context.Context, session *Session) (string, error)
	// DeleteSessionsForUser deletes all sessions for the user with the given ID.
	DeleteSessionsForUser(ctx context.Context, userID string) error
	// GetSessionByID retrieves the session with the given ID.
	GetSessionByID(ctx context.Context, sessionID string) (*Session, error)
}

//go:generate mockery --name=Manager --case=underscore --inpackage
type Manager interface {
	// CreateSession creates a new session for the user with the given ID.
	CreateSession(ctx context.Context, userID string) (*Session, error)
	// ValidateSession checks if the session is valid.
	ValidateSession(ctx context.Context, sessionID string) (*Session, error)
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
	err := m.repo.DeleteSessionsForUser(ctx, userID)
	if err != nil {
		return &Session{}, err
	}
	session := NewSession(userID)
	_, err = m.repo.CreateSession(ctx, session)
	if err != nil {
		return &Session{}, err
	}
	return session, nil
}

func (m *manager) ValidateSession(ctx context.Context, sessionID string) (*Session, error) {
	session, err := m.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return &Session{}, err
	}
	if !session.IsValid() {
		return &Session{}, ErrInvalidSession
	}
	return session, nil
}
