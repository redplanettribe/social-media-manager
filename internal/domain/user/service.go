package user

import (
	"context"

	"github.com/google/uuid"
)

//go:generate mockery --name=Service --case=underscore --inpackage
type Service interface {
	CreateUser(ctx context.Context, username, password, email string) error
	GetUser(ctx context.Context, id string) (*UserResponse, error)
	// Additional methods as needed
}

// Create a concrete implementation
type service struct {
	repo     Repository
	password PasswordHasher
}

// Update the constructor to return the interface
func NewService(repo Repository, passwordHasher PasswordHasher) Service {
	return &service{
		repo:     repo,
		password: passwordHasher,
	}
}

func (s *service) CreateUser(ctx context.Context, username, password, email string) error {
	existingUser, err := s.repo.FindByUsernameOrEmail(ctx, username, email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return ErrExistingUser
	}

	hashedPassword, salt, err := s.password.Hash(password)
	if err != nil {
		return err
	}

	user, err := NewUser(username, hashedPassword, salt, email)
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
