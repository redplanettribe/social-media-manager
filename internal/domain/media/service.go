package media

import (
	"context"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
)

type Service interface {
	UploadMedia(ctx context.Context, projectID, postID, fileName string, data []byte) (*MetaData, error)
}

type service struct {
	repo       Repository
	objectRepo ObjectRepository
}

func NewService(repo Repository, objectRepo ObjectRepository) Service {
	return &service{
		repo:       repo,
		objectRepo: objectRepo,
	}
}

func (s *service) UploadMedia(ctx context.Context, projectID, postID, fileName string, data []byte) (*MetaData, error) {
	userID := ctx.Value(middlewares.UserIDKey).(string)

    metadata, err := NewMetadata(postID, userID, fileName, data)
	if err != nil {
		return nil, err
	}

	err = s.objectRepo.UploadFile(ctx, projectID, postID, fileName, data, metadata)
	if err != nil {
		return nil, err
	}

    // TODO:
    // we should process the file and save the thumbnail also

	return s.repo.SaveMetadata(ctx, metadata)
}
