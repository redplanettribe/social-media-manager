package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/session"
)

//go:generate mockery --name=Service --case=underscore --inpackage
type Service interface {
	CreateUser(ctx context.Context, username, password, email string) error
	GetUser(ctx context.Context, id string) (*UserResponse, error)
	Signin(ctx context.Context, email, password string) error
	// Additional methods as needed
}

// Create a concrete implementation
type service struct {
	repo     Repository
	password PasswordHasher
	session  session.Manager
}

// Update the constructor to return the interface
func NewService(repo Repository, session session.Manager, passwordHasher PasswordHasher) Service {
	return &service{
		repo:     repo,
		password: passwordHasher,
		session:  session,
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
	return s.repo.FindByID(ctx, userID.String())
}

func (s *service) Signin(ctx context.Context, email, password string) error {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	if !s.password.Validate(password, user.HashedPasword, user.Salt) {
		return ErrInvalidPassword
	}
	// creates a new session
	_, err = s.session.CreateSession(user.ID)
	if err != nil {
		return err
	}
	return nil
}
