package user

import "context"

type Repository interface {
	Save(ctx context.Context, usr *User) error
	FindByID(ctx context.Context, id UserID) (*UserResponse, error)
	// Additional methods as needed
}
