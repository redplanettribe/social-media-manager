package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/session"
)

type SessionRepository struct {
	db *pgx.Conn
}

func NewSessionRepository(db *pgx.Conn) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) CreateSession(ctx context.Context, session *session.Session) (string, error) {
	query := `
		INSERT INTO sessions (id, user_id, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(ctx, query, session.ID, session.UserID, session.CreatedAt, session.ExpiresAt)
	if err != nil {
		return "", err
	}
	return session.ID, nil
}

func (r *SessionRepository) DeleteSessionsForUser(ctx context.Context, userID string) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}
	return nil
}

func (r *SessionRepository) GetSessionByID(ctx context.Context, sessionID string) (*session.Session, error) {
	query := `
		SELECT id, user_id, created_at, expires_at
		FROM sessions
		WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, sessionID)
	var s session.Session
	err := row.Scan(&s.ID, &s.UserID, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
