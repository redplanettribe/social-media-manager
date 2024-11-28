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

func (r *UserRepository) FindByID(ctx context.Context, id user.UserID) (*user.UserResponse, error) {
	query := `
		SELECT id, username, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id.String())
	u := &user.UserResponse{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}
