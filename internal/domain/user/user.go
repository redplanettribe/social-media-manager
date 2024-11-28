package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type UserID uuid.UUID

func (id UserID) String() string {
	return uuid.UUID(id).String()
}

type User struct {
	ID          UserID
	Username    string
	PaswordHash string
	Email       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewUser(username, hashedPw, email string) (*User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if hashedPw == "" {
		return nil, errors.New("password cannot be empty")
	}
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	return &User{
		ID:          UserID(uuid.New()),
		Username:    username,
		PaswordHash: hashedPw,
		Email:       email,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}
