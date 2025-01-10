package media

import (
	"context"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	UploadMedia(ctx context.Context, projectID, postID, fileName string, data []byte) (*MetaData, error)
	GetMediaFile(ctx context.Context, projectID, postID, fileName string) (*Media, error)
	GetMediaForPost(ctx context.Context, projectID, postID, platformID string) ([]*Media, error)
	LinkMediaToPublishPost(ctx context.Context, projectID, postID, mediaID, platformID string) error
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

	existingMetadata, err := s.repo.GetMetadata(ctx, postID, fileName)
	if err == nil && existingMetadata != nil {
		return nil, ErrFileAlreadyExists
	}

	md, err := NewMetadata(postID, userID, fileName, data)
	if err != nil {
		return nil, err
	}
	err = s.objectRepo.UploadFile(ctx, projectID, postID, fileName, data, md)
	if err != nil {
		return nil, err
	}

	// TODO:
	// we should process the file and save the thumbnail also

	return s.repo.SaveMetadata(ctx, md)
}

func (s *service) GetMediaFile(ctx context.Context, projectID, postID, fileName string) (*Media, error) {
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
		return nil, err
	}

	return &Media{
		Data:     file,
		MetaData: metadata,
	}, nil
}

func (s *service) GetMediaForPost(ctx context.Context, projectID, postID, platformID string) ([]*Media, error) {
	mediaNames, err := s.repo.GetMediaFileNamesForPost(ctx, postID, platformID)
	if err != nil {
		return nil, err
	}
	var (
		medias  = make([]*Media, len(mediaNames))
		g, gCtx = errgroup.WithContext(ctx)
	)

	for i, mediaName := range mediaNames {
		i, name := i, mediaName
		g.Go(func() error {
			media, err := s.GetMediaFile(gCtx, projectID, postID, name)
			if err != nil {
				return err
			}
			medias[i] = media
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return medias, nil
}

func (s *service) LinkMediaToPublishPost(ctx context.Context, projectID, postID, mediaID, platformID string) error {

	var (
		doesPostBelongToProject    bool
		doesMediaBelongToPost     bool
		isThePostLinkedToPlatform bool
		isPlatformEnabled         bool
		isAlreadyLinked           bool
	)
	
	g, gCtx := errgroup.WithContext(ctx)
	
	g.Go(func() error {
		var err error
		doesPostBelongToProject, err = s.repo.DoesPostBelongToProject(gCtx, projectID, postID)
		return err
	})
	
	g.Go(func() error {
		var err error
		doesMediaBelongToPost, err = s.repo.DoesMediaBelongToPost(gCtx, postID, mediaID)
		return err
	})
	
	g.Go(func() error {
		var err error
		isPlatformEnabled, err = s.repo.IsPlatformEnabledForProject(gCtx, projectID, platformID)
		return err
	})

	g.Go(func() error {
		var err error
		isThePostLinkedToPlatform, err = s.repo.IsThePostEnabledToPlatform(gCtx, postID, platformID)
		return err
	})

	g.Go(func() error {
		var err error
		isAlreadyLinked, err = s.repo.IsMediaLinkedToPublishPost(gCtx, postID, mediaID, platformID)
		return err
	})
	
	if err := g.Wait(); err != nil {
		return err
	}
	
	if !doesPostBelongToProject {
		return ErrPostDoesNotBelongToProject
	}
	if !doesMediaBelongToPost {
		return ErrMediaDoesNotBelongToPost
	}
	if !isPlatformEnabled {
		return ErrPlatformNotEnabledForProject
	}
	if !isThePostLinkedToPlatform {
		return ErrPostNotLinkedToPlatform
	}
	if isAlreadyLinked {
		return ErrMediaAlreadyLinkedToPost
	}

	return s.repo.LinkMediaToPublishPost(ctx, postID, mediaID, platformID)
}
