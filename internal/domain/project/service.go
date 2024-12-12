package project

import (
	"context"
	"errors"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
)

type Service interface {
	CreateProject(ctx context.Context, name, description string) (*Project, error)
	ListProjects(ctx context.Context) ([]*Project, error)
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
