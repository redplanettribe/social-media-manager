package project

import (
	"context"
	"errors"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	CreateProject(ctx context.Context, name, description string) (*Project, error)
	ListProjects(ctx context.Context) ([]*Project, error)
	GetUserRoles(ctx context.Context, userID, projectID string) ([]string, error)
	GetProject(ctx context.Context, projectID string) (*ProjectResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateProject(ctx context.Context, name, description string) (*Project, error) {
	userID, ok := ctx.Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, errors.New("userID not found in context")
	}

	project, err := NewProject(name, description, userID)
	if err != nil {
		return nil, err
	}

	project, err = s.repo.Save(ctx, project)
	if err != nil {
		return nil, err
	}

	err = s.repo.AssignProjectOwner(ctx, project.ID, userID)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *service) ListProjects(ctx context.Context) ([]*Project, error) {
	userID, ok := ctx.Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		return nil, errors.New("userID not found in context")
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
		g 	 errgroup.Group
	)

	g.Go(func() error {
		var err error
		project, err = s.repo.GetProject(ctx, projectID)
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