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
	EnableSocialPlatform(ctx context.Context, projectID, socialPlatformID string) error
	DoesSocialPlatformExist(ctx context.Context, socialPlatformID string) (bool, error)
	IsProjectSocialPlatformEnabled(ctx context.Context, projectID, socialPlatformID string) (bool, error)
	GetEnabledSocialPlatforms(ctx context.Context, projectID string) ([]SocialPlatform, error)
	GetProjectSchedule(ctx context.Context, projectID string) (*WeeklyPostSchedule, error)
	SaveSchedule(ctx context.Context, projectID string, schedule *WeeklyPostSchedule) error
	CreateProjectSettings(ctx context.Context, projectID string, schedule *WeeklyPostSchedule) error
	FindActiveProjectsChunk(ctx context.Context, limit, offset int) ([]*Project, error)
	SetDefaultUser(ctx context.Context, projectID, userID string) error
	GetDefaultUserID(ctx context.Context, projectID string) (string, error)
	GetPlatformInfo(ctx context.Context, userID, platformID string) (*UserPlatformInfo, error)
}
