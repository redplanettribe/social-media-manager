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
		INSERT INTO users (id, username, password_hash, salt, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, u.ID, u.Username, u.PaswordHash, u.Salt, u.Email, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.UserResponse, error) {
	query := `
		SELECT id, username, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)
	u := &user.UserResponse{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) FindByUsernameOrEmail(ctx context.Context, username, email string) (*user.UserResponse, error) {
	query := `
		SELECT id, username, email, created_at, updated_at
		FROM users
		WHERE username = $1 OR email = $2
	`
	row := r.db.QueryRow(ctx, query, username, email)
	u := &user.UserResponse{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	} else if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.FullUserResponse, error) {
	query := `
		SELECT id, username, email, password_hash, salt, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	row := r.db.QueryRow(ctx, query, email)
	u := &user.FullUserResponse{}
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.HashedPasword, &u.Salt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	} else if err == pgx.ErrNoRows {
		return nil, nil
	}
	return u, nil
}

func (r *UserRepository) GetRoles(ctx context.Context) ([]*user.Role, error) {
	query := `
		SELECT id, role
		FROM roles
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := []*user.Role{}
	for rows.Next() {
		r := &user.Role{}
		err := rows.Scan(&r.ID, &r.Name)
		if err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, nil
}
