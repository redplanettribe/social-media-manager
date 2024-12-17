package postgres

import (
	"context"

	"github.com/pedrodcsjostrom/opencm/internal/domain/user"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) (*user.UserResponse, error) {
	query := `
		INSERT INTO users (id, username, password_hash, salt, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, u.ID, u.Username, u.PaswordHash, u.Salt, u.Email, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return &user.UserResponse{}, err
	}
	return &user.UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
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

func (r *UserRepository) FindByIDWithRoles(ctx context.Context, id string) (*user.UserResponse, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.updated_at, r.id, r.role
		FROM users u
		LEFT JOIN user_roles ur ON u.id = ur.user_id
		LEFT JOIN roles r ON ur.role_id = r.id
		WHERE u.id = $1
	`
	rows, err := r.db.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	u := &user.UserResponse{}
	for rows.Next() {
		var role user.AppRole
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt, &u.UpdatedAt, &role.ID, &role.Name)
		if err != nil {
			return nil, err
		}
		u.Roles = append(u.Roles, role)
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

func (r *UserRepository) GetRoles(ctx context.Context) (*[]user.AppRole, error) {
	query := `
		SELECT id, role
		FROM roles
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := []user.AppRole{}
	for rows.Next() {
		var role user.AppRole
		err := rows.Scan(&role.ID, &role.Name)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return &roles, nil
}

func (r *UserRepository) GetUserRoles(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT r.id, r.role
		FROM roles r
		LEFT JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := []string{}
	for rows.Next() {
		var role user.AppRole
		err := rows.Scan(&role.ID, &role.Name)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role.Name)
	}
	return roles, nil
}

func (r *UserRepository) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
	query := `
        INSERT INTO user_roles (user_id, role_id)
        VALUES ($1, $2)
        ON CONFLICT (user_id, role_id) DO NOTHING
    `
	_, err := r.db.Exec(ctx, query, userID, roleID)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) AssignDefaultRoleToUser(ctx context.Context, userID string) error {
	query := `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, userID, 1)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID string) error {
	query := `
		DELETE FROM user_roles
		WHERE user_id = $1 AND role_id = $2
	`
	_, err := r.db.Exec(ctx, query, userID, roleID)
	if err != nil {
		return err
	}
	return nil
}
