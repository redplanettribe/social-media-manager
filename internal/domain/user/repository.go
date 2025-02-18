package user

import "context"

//go:generate mockery --name=Repository --config=../../../mockery.yaml
type Repository interface {
	Save(ctx context.Context, usr *User) (*UserResponse, error)
	FindByID(ctx context.Context, id string) (*UserResponse, error)
	FindByIDWithRoles(ctx context.Context, id string) (*UserResponse, error)
	FindByUsernameOrEmail(ctx context.Context, username, email string) (*UserResponse, error)
	FindByEmail(ctx context.Context, email string) (*FullUserResponse, error)
	GetRoles(ctx context.Context) (*[]AppRole, error)
	GetUserRoles(ctx context.Context, userID string) ([]string, error)
	AssignRoleToUser(ctx context.Context, userID, roleID string) error
	AssignDefaultRoleToUser(ctx context.Context, userID string) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID string) error
	// Additional methods as needed
}
