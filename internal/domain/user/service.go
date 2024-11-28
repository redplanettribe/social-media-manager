package user

import (
	"context"

	"github.com/google/uuid"
)

// Instead of having an interface type
type Service interface {
	CreateUser(ctx context.Context, username, password, email string) error
	GetUser(ctx context.Context, id string) (*UserResponse, error)
	// Additional methods as needed
}

// Create a concrete implementation
type service struct {
	repo Repository
}

// Update the constructor to return the interface
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateUser(ctx context.Context, username, password, email string) error {
	hashedPassword := password + "hashed"
	user, err := NewUser(username, hashedPassword, email)
	if err != nil {
		return err
	}
	return s.repo.Save(ctx, user)
}

func (s *service) GetUser(ctx context.Context, id string) (*UserResponse, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, UserID(userID))
}
