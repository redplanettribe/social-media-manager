package user

import "context"

type Repository interface {
	Save(ctx context.Context, usr *User) error
	// FindByID(id UserID) (*User, error)
	// Additional methods as needed
}
