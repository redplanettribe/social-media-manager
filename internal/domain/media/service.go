package media

import (
	"context"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	UploadMedia(ctx context.Context, projectID, postID, fileName string, data []byte) (*MetaData, error)
	GetMediaFile(ctx context.Context, projectID, postID, fileName string) ([]byte, *MetaData, error)
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

func (s *service) GetMediaFile(ctx context.Context, projectID, postID, fileName string) ([]byte, *MetaData, error) {
	var (
		file     []byte
		metadata *MetaData
		eg       errgroup.Group
	)

	eg.Go(func() error {
		var err error
		file, err = s.objectRepo.GetFile(ctx, projectID, postID, fileName)
		return err
	})

	eg.Go(func() error {
		var err error
		metadata, err = s.repo.GetMetadata(ctx, postID, fileName)
		return err
	})

	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}

	return file, metadata, nil
}
