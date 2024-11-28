package postgres

import (
	"context"

	"github.com/pedrodcsjostrom/opencm/internal/domain/user"

	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *pgx.Conn
}

func NewUserRepository(db *pgx.Conn) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (id, username, password_hash, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query, u.ID.String(), u.Username, u.PaswordHash, u.Email, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}
