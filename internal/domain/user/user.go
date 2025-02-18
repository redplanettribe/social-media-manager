package user

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/redplanettribe/social-media-manager/internal/infrastructure/session"
)

var (
	ErrExistingUser                 = errors.New("email already registered")
	ErrUserNotFound                 = errors.New("user not found")
	ErrNoUserInContext              = errors.New("no user in context")
	ErrInvalidPassword              = errors.New("invalid password")
	ErrNoDeviceFingerprintInContext = errors.New("no device fingerprint in context")
)

type User struct {
	ID          string
	Username    string
	PaswordHash string
	Salt        string
	Email       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Roles     []AppRole `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginResponse struct {
	User    *UserResponse
	Session *session.Session
}

type FullUserResponse struct {
	ID            string    `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	HashedPasword string    `json:"password"`
	Salt          string    `json:"salt"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AppRole struct {
	ID   string
	Name string
}

func NewUser(username, hashedPw, salt, email string) (*User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if hashedPw == "" {
		return nil, errors.New("password cannot be empty")
	}
	if salt == "" {
		return nil, errors.New("salt cannot be empty")
	}
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	return &User{
		ID:          uuid.New().String(),
		Username:    username,
		PaswordHash: hashedPw,
		Salt:        salt,
		Email:       email,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}, nil
}

func sanitize(fu *FullUserResponse) *UserResponse {
	return &UserResponse{
		ID:        fu.ID,
		Username:  fu.Username,
		Email:     fu.Email,
		CreatedAt: fu.CreatedAt,
		UpdatedAt: fu.UpdatedAt,
	}
}
