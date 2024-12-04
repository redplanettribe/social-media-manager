package project

import (
	"context"
	"errors"
	"fmt"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
)

type Service interface {
	CreateProject(ctx context.Context, name, description string) (*Project, error)
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
	userIDInterface := ctx.Value(middlewares.UserIDKey)
	if userIDInterface == nil {
		fmt.Println("userID not found in context, Context: ", ctx)
		return &Project{}, errors.New("userID not found in context")
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		return &Project{}, errors.New("userID in context is not a string")
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
