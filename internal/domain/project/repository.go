package project

import "context"

type Repository interface {
	Save(ctx context.Context, project *Project) (*Project, error)
	AssignProjectOwner(ctx context.Context, projectID, userID string) error
	ListByUserID(ctx context.Context, userID string) ([]*Project, error)
	GetUserRoles(ctx context.Context, userID, projectID string) ([]string, error)
	FindProjectByID(ctx context.Context, projectID string) (*Project, error)
	GetProjectUsers(ctx context.Context, projectID string) ([]*TeamMember, error)
	AddUserToProject(ctx context.Context, projectID, userID string) error
	IsUserInProject(ctx context.Context, projectID, userID string) (bool, error)
	DoesProjectNameExist(ctx context.Context, name, userID string) (bool, error)
}
