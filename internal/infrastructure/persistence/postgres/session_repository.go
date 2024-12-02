package postgres

import "github.com/jackc/pgx/v5"

type SessionRepository struct {
	db *pgx.Conn
}

func NewSessionRepository(db *pgx.Conn) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateSession(string) (string, error) {
	// To be implemented
	return "", nil
}

func (r *SessionRepository) FindSession(sessionID string) (string, error) {
	// To be implemented
	return "", nil
}

func (r *SessionRepository) DeleteSession(sessionID string) error {
	// To be implemented
	return nil
}
