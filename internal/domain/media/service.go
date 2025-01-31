package media

import (
	"context"
	"errors"
	"fmt"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	UploadMedia(ctx context.Context, projectID, postID, fileName, altText string, data []byte) (DownloadMetaData, error)
	GetDownloadMetaData(ctx context.Context, projectID, postID, fileName string) (DownloadMetaData, error)
	GetMediaFile(ctx context.Context, projectID, postID, fileName string) (*Media, error)
	GetMediaForPublishPost(ctx context.Context, projectID, postID, platformID string) ([]*Media, error)
	GetDownloadMetadataForPublishPost(ctx context.Context, projectID, postID, platformID string) ([]*DownloadMetaData, error)
	LinkMediaToPublishPost(ctx context.Context, projectID, postID, mediaID, platformID string) error
	GetDownloadMetadataDataForPost(ctx context.Context, projectID, postID string) ([]*DownloadMetaData, error)
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

func (s *service) UploadMedia(ctx context.Context, projectID, postID, fileName, altText string, data []byte) (DownloadMetaData, error) {
	userID := ctx.Value(middlewares.UserIDKey).(string)

	existingMetadata, err := s.repo.GetMetadata(ctx, postID, fileName)
	if err == nil && existingMetadata != nil {
		return DownloadMetaData{}, ErrFileAlreadyExists
	}

	processor, err := GetProcessor(fileName)
	if err != nil {
		return DownloadMetaData{}, err
	}

	var (
		g          errgroup.Group
		mediaInfo  *MediaInfo
		tMediaInfo *MediaInfo
		thumnail   *[]byte
		md         *MetaData
		tmd        *MetaData
	)

	// Analyze the media and get the thumbnail with its info
	g.Go(func() error {
		var err error
		mediaInfo, err = processor.Analyze(data)
		return err
	})
	g.Go(func() error {
		var err error
		if processor.GetMediaType() == MediaTypeDocument {
			return nil
		}
		thumnail, err = processor.GetThumbnail(data)
		if err != nil {
			return err
		}

		tMediaInfo, err = processor.Analyze(*thumnail)
		return err
	})

	if err := g.Wait(); err != nil {
		return DownloadMetaData{}, errors.Join(errors.New("failed to analyze media"), err)
	}

	// Upload the media and thumbnail, save the metadata
	var mediaUrl string
	g.Go(func() error {
		var err error
		md, err = NewMetadata(postID, userID, fileName, altText, data, mediaInfo)
		if err != nil {
			return err
		}
		err = s.objectRepo.UploadFile(ctx, projectID, postID, fileName, data, md)
		if err != nil {
			return err
		}
		mediaUrl, err = s.objectRepo.GetSignedURL(ctx, projectID, postID, fileName)
		if err != nil {
			return err
		}
		_, err = s.repo.SaveMetadata(ctx, md)
		return err
	})

	var thumbnailUrl string
	g.Go(func() error {
		var err error
		if processor.GetMediaType() == MediaTypeDocument {
			return nil
		}
		thumbnailFileName := getThumbnailName(fileName)
		tmd, err = NewMetadata(postID, userID, thumbnailFileName, altText, *thumnail, tMediaInfo)
		if err != nil {
			return err
		}
		err = s.objectRepo.UploadFile(ctx, projectID, postID, thumbnailFileName, *thumnail, tmd)
		if err != nil {
			return err
		}
		thumbnailUrl, err = s.objectRepo.GetSignedURL(ctx, projectID, postID, thumbnailFileName)
		if err != nil {
			return err
		}
		_, err = s.repo.SaveMetadata(ctx, tmd)
		return err
	})

	if err = g.Wait(); err != nil {
		return DownloadMetaData{}, err
	}

	fmt.Println("mediaUrl", mediaUrl)
	fmt.Println("thumbnailUrl", thumbnailUrl)

	return DownloadMetaData{
		Url:          &mediaUrl,
		UrlThumbnail: &thumbnailUrl,
		MetaData:     md,
	}, nil
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

func (s *service) GetMediaForPublishPost(ctx context.Context, projectID, postID, platformID string) ([]*Media, error) {
	filenames, err := s.repo.GetMediaFileNamesForPublishPost(ctx, postID, platformID)
	if err != nil {
		return nil, err
	}
	var (
		medias  = make([]*Media, len(filenames))
		g, gCtx = errgroup.WithContext(ctx)
	)

	for i, mediaName := range filenames {
		i, name := i, mediaName
		g.Go(func() error {
			media, err := s.GetMediaFile(gCtx, projectID, postID, name)
			if err != nil {
				return err
			}
			if media.MetaData.IsVideo() {
				thumbnailName := getThumbnailName(name)
				thumbnail, err := s.GetMediaFile(gCtx, projectID, postID, thumbnailName)
				if err != nil {
					return err
				}
				media.Thumbnail = thumbnail
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

func (s *service) GetDownloadMetadataForPublishPost(ctx context.Context, projectID, postID, platformID string) ([]*DownloadMetaData, error) {
	filenames, err := s.repo.GetMediaFileNamesForPublishPost(ctx, postID, platformID)
	if err != nil {
		return nil, err
	}
	var (
		downloadMetaDatas = make([]*DownloadMetaData, len(filenames))
		g, gCtx           = errgroup.WithContext(ctx)
	)

	for i, filename := range filenames {
		i, filename := i, filename
		g.Go(func() error {
			downloadMetaData, err := s.GetDownloadMetaData(gCtx, projectID, postID, filename)
			if err != nil {
				return err
			}
			downloadMetaDatas[i] = &downloadMetaData
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return downloadMetaDatas, nil
}

func (s *service) LinkMediaToPublishPost(ctx context.Context, projectID, postID, mediaID, platformID string) error {

	var (
		doesPostBelongToProject   bool
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

func (s *service) GetDownloadMetaData(ctx context.Context, projectID, postID, fileName string) (DownloadMetaData, error) {
	var (
		mediaUrl      string
		thumbnailUrl  string
		metadata      *MetaData
		thumbnailName string
		eg            errgroup.Group
	)

	eg.Go(func() error {
		var err error
		mediaUrl, err = s.objectRepo.GetSignedURL(ctx, projectID, postID, fileName)
		return err
	})

	eg.Go(func() error {
		var err error
		metadata, err = s.repo.GetMetadata(ctx, postID, fileName)
		return err
	})

	eg.Go(func() error {
		var err error
		thumbnailName = getThumbnailName(fileName)
		thumbnailUrl, err = s.objectRepo.GetSignedURL(ctx, projectID, postID, thumbnailName)
		if err != nil {
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return DownloadMetaData{}, err
	}

	return DownloadMetaData{
		Url:          &mediaUrl,
		UrlThumbnail: &thumbnailUrl,
		MetaData:     metadata,
	}, nil
}

func (s *service) GetDownloadMetadataDataForPost(ctx context.Context, projectID string, postID string) ([]*DownloadMetaData, error) {
	filenames, err := s.repo.ListMediaFilesForPost(ctx, postID)
	if err != nil {
		return nil, err
	}

	var (
		downloadMetaDatas = make([]*DownloadMetaData, len(filenames))
		g, gCtx           = errgroup.WithContext(ctx)
	)

	for i, filename := range filenames {
		i, filename := i, filename
		g.Go(func() error {
			downloadMetaData, err := s.GetDownloadMetaData(gCtx, projectID, postID, filename)
			if err != nil {
				return err
			}
			downloadMetaDatas[i] = &downloadMetaData
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return downloadMetaDatas, nil
}
