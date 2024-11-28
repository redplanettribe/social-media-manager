package user

import "context"

// Instead of having an interface type
type Service interface {
	CreateUser(ctx context.Context, username, password, email string) error
	// Additional methods as needed
}

// Create a concrete implementation
type serviceImpl struct {
	repo Repository
}

// Update the constructor to return the interface
func NewService(repo Repository) Service {
	return &serviceImpl{repo: repo}
}

// Implement the CreateUser method on the concrete type
func (s *serviceImpl) CreateUser(ctx context.Context, username, password, email string) error {
	hashedPassword := password + "hashed"
	user, err := NewUser(username, hashedPassword, email)
	if err != nil {
		return err
	}
	return s.repo.Save(ctx, user)
}
