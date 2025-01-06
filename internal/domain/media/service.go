package media

import (
	"context"
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
	metadata, err := NewMetadata(postID, fileName, data)
	if err != nil {
		return nil, err
	}
	err = s.objectRepo.UploadFile(ctx, projectID, postID, fileName, data)
	if err != nil {
		return nil, err
	}

	return s.repo.SaveMetadata(ctx, metadata)
}

