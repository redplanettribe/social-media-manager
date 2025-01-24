package project

import (
	"context"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/user"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
	"golang.org/x/sync/errgroup"
)

//go:generate mockery --name=Service --case=underscore --inpackage
type Service interface {
	CreateProject(ctx context.Context, name, description string) (*Project, error)
	ListProjects(ctx context.Context) ([]*Project, error)
	GetUserRoles(ctx context.Context, userID, projectID string) ([]string, error)
	GetProject(ctx context.Context, projectID string) (*ProjectResponse, error)
	AddUserToProject(ctx context.Context, projectID, email string) error
	EnableSocialPlatform(ctx context.Context, projectID, socialPlatformID string) error
	GetEnabledSocialPlatforms(ctx context.Context, projectID string) ([]SocialPlatform, error)
	SetTimeZone(ctx context.Context, projectID, timeZone string) error
	AddTimeSlot(ctx context.Context, projectID string, dayOfWeek time.Weekday, hour, minute int) error
	IsProjectTimeToPublish(ctx context.Context, projectID string) (bool, error)
	FindActiveProjectsChunk(ctx context.Context, offset, chunkSize int) ([]*Project, error)
	SetDefaultUser(ctx context.Context, projectID, userID string) error
	GetDefaultUserPlatformInfo(ctx context.Context, projecID, platformID string) (*UserPlatformInfo, error)
}

type service struct {
	repo     Repository
	userRepo user.Repository
}

func NewService(repo Repository, uRepo user.Repository) Service {
	return &service{
		repo:     repo,
		userRepo: uRepo,
	}
}

func (s *service) CreateProject(ctx context.Context, name, description string) (*Project, error) {
	userID, ok := ctx.Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, ErrNoUserIDInContext
	}

	if ok, err := s.repo.DoesProjectNameExist(ctx, name, userID); err != nil {
		return nil, err
	} else if ok {
		return nil, ErrProjectExists
	}

	project, err := NewProject(name, description, userID)
	if err != nil {
		return nil, err
	}

	project, err = s.repo.Save(ctx, project)
	if err != nil {
		return nil, err
	}

	shc := NewWeeklyPostSchedule("America/New_York", []TimeSlot{}) // TODO: get timezone from request
	err = s.repo.CreateProjectSettings(ctx, project.ID, shc)
	if err != nil {
		return nil, err
	}

	err = s.repo.AssignProjectOwner(ctx, project.ID, userID)
	if err != nil {
		return nil, err
	}

	err = s.repo.SetDefaultUser(ctx, project.ID, userID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *service) ListProjects(ctx context.Context) ([]*Project, error) {
	userID, ok := ctx.Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, ErrNoUserIDInContext
	}

	projects, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (s *service) GetUserRoles(ctx context.Context, userID, projectID string) ([]string, error) {
	return s.repo.GetUserRoles(ctx, userID, projectID)
}

func (s *service) GetProject(ctx context.Context, projectID string) (*ProjectResponse, error) {
	var (
		project *Project
		users   []*TeamMember
		g       errgroup.Group
	)

	g.Go(func() error {
		var err error
		project, err = s.repo.FindProjectByID(ctx, projectID)
		return err
	})

	g.Go(func() error {
		var err error
		users, err = s.repo.GetProjectUsers(ctx, projectID)
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &ProjectResponse{
		Project: project,
		Users:   users,
	}, nil
}

func (s *service) AddUserToProject(ctx context.Context, projectID, email string) error {
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if u == nil {
		return ErrUserNotFound
	}

	userID := u.ID
	if ok, err := s.repo.IsUserInProject(ctx, projectID, userID); err != nil {
		return err
	} else if ok {
		return ErrUserAlreadyInProject
	}

	return s.repo.AddUserToProject(ctx, projectID, userID)
}

func (s *service) EnableSocialPlatform(ctx context.Context, projectID, socialPlatformID string) error {
	var (
		exists  bool
		enabled bool
		g       errgroup.Group
	)

	g.Go(func() error {
		var err error
		exists, err = s.repo.DoesSocialPlatformExist(ctx, socialPlatformID)
		return err
	})

	g.Go(func() error {
		var err error
		enabled, err = s.repo.IsProjectSocialPlatformEnabled(ctx, projectID, socialPlatformID)
		return err
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if !exists {
		return ErrSocialPlatformNotFound
	}

	if enabled {
		return ErrSocialPlatformAlreadyEnabled
	}

	return s.repo.EnableSocialPlatform(ctx, projectID, socialPlatformID)
}

func (s *service) GetEnabledSocialPlatforms(ctx context.Context, projectID string) ([]SocialPlatform, error) {
	return s.repo.GetEnabledSocialPlatforms(ctx, projectID)
}

func (s *service) SetTimeZone(ctx context.Context, projectID, timeZone string) error {
	sch, err := s.repo.GetProjectSchedule(ctx, projectID)
	if err != nil {
		return err
	}
	sch.SetTimeZone(timeZone)
	err = s.repo.SaveSchedule(ctx, projectID, sch)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) AddTimeSlot(ctx context.Context, projectID string, dayOfWeek time.Weekday, hour, minute int) error {
	sch, err := s.repo.GetProjectSchedule(ctx, projectID)
	if err != nil {
		return err
	}
	err = sch.AddSlot(dayOfWeek, hour, minute)
	if err != nil {
		return err
	}
	err = s.repo.SaveSchedule(ctx, projectID, sch)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) IsProjectTimeToPublish(ctx context.Context, projectID string) (bool, error) {
	sch, err := s.repo.GetProjectSchedule(ctx, projectID)
	if err != nil {
		return false, err
	}
	return sch.IsTime(time.Now()), nil
}

func (s *service) FindActiveProjectsChunk(ctx context.Context, offset, chunkSize int) ([]*Project, error) {
	return s.repo.FindActiveProjectsChunk(ctx, offset, chunkSize)
}

func (s *service) SetDefaultUser(ctx context.Context, projectID, userID string) error {
	ok, err := s.repo.IsUserInProject(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUserNotInProject
	}
	return s.repo.SetDefaultUser(ctx, projectID, userID)
}

func (s *service) GetDefaultUserPlatformInfo(ctx context.Context, projectID, platformID string) (*UserPlatformInfo, error) {
	defaultUserID, err := s.repo.GetDefaultUserID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	pInfo, err := s.repo.GetPlatformInfo(ctx, defaultUserID, platformID)
	if err != nil {
		return nil, err
	}
	return pInfo, nil
}
