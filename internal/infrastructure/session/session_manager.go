package session

//go:generate mockery --name=Manager --case=underscore --inpackage
type Manager interface {
	// CreateSession creates a new session for the user with the given ID.
	CreateSession(userID string) (string, error)
	// ValidateSession checks if the session is valid.
	ValidateSession(sessionID string) (string, error)
	// InvalidateSession invalidates the session.
	InvalidateSession(sessionID string) error
	// InvalidateAllSessions invalidates all sessions for the user with the given ID.
	InvalidateAllSessions(userID string) error
}

//go:generate mockery --name=Repository --case=underscore --inpackage
type Repository interface {
	// CreateSession creates a new session for the user with the given ID.
	CreateSession(userID string) (string, error)
	// FindSession returns the session with the given ID.
	FindSession(sessionID string) (string, error)
	// DeleteSession deletes the session with the given ID.
	DeleteSession(sessionID string) error
}

type manager struct {
	repo Repository
}

func NewManager(repo Repository) Manager {
	return &manager{
		repo: repo,
	}
}

func (m *manager) InvalidateAllSessions(userID string) error {
	// To be implemented
	return nil
}

func (m *manager) CreateSession(userID string) (string, error) {
	// To be implemented
	return "", nil
}

func (m *manager) ValidateSession(sessionID string) (string, error) {
	// To be implemented
	return "", nil
}

func (m *manager) InvalidateSession(sessionID string) error {
	// To be implemented
	return nil
}
