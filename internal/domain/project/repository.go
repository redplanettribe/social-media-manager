package project

import "context"

type Repository interface {
	Save(ctx context.Context, project *Project) (*Project, error)
	AssignProjectOwner(ctx context.Context, projectID, userID string) error
	ListByUserID(ctx context.Context, userID string) ([]*Project, error)
	GetUserRoles(ctx context.Context, userID, projectID string) ([]string, error)
	GetProject(ctx context.Context, projectID string) (*Project, error)
}
