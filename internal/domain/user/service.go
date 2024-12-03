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
	Login(ctx context.Context, email, password string) (*session.Session, error)
	GetRoles(ctx context.Context) ([]*Role, error)
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

func (s *service) Login(ctx context.Context, email, password string) (*session.Session, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return &session.Session{}, err
	}
	if user == nil {
		return &session.Session{}, ErrUserNotFound
	}
	if !s.password.Validate(password, user.HashedPasword, user.Salt) {
		return &session.Session{}, ErrInvalidPassword
	}
	return s.session.CreateSession(ctx, user.ID)
}

func (s *service) Logout(ctx context.Context, sessionID string) error {
	// TODO: Implement the logout logic
	return nil
}

func (s *service) UpdatePassword(ctx context.Context, userID, password string) error {
	// TODO: Implement the update password logic
	return nil
}

func (s *service) UpdateEmail(ctx context.Context, userID, email string) error {
	// TODO: Implement the update email logic
	return nil
}

func (s *service) GetRoles(ctx context.Context) ([]*Role, error) {
	return s.repo.GetRoles(ctx)
}

// func (s *service) AssignRoleToUser(ctx context.Context, userID, roleID string) error {
// 	err := s.repo.AssignRoleToUser(ctx, userID, roleID)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (s *service) RevokeRoleFromUser(ctx context.Context, userID, roleID string) error {
// 	err := s.repo.RevokeRoleFromUser(ctx, userID, roleID)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
