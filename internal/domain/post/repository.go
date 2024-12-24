package post

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context,post *Post) error
	FindByID(ctx context.Context,id string) (*Post, error)
	FindByProjectID(ctx context.Context,teamID uuid.UUID) ([]*Post, error)
}
