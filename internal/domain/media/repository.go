package media

import "context"

type Repository interface {
	SaveMetadata(ctx context.Context, media *MetaData) (*MetaData, error)
	GetMetadata(ctx context.Context, postID, fileName string) (*MetaData, error)
	GetMediaFileNamesForPublishPost(ctx context.Context, postID, platformID string) ([]string, error)
	LinkMediaToPublishPost(ctx context.Context, postID, fileName, platformID string) error
	UnlinkMediaFromPublishPost(ctx context.Context, postID, fileName, platformID string) error
	DoesPostBelongToProject(ctx context.Context, projectID, postID string) (bool, error)
	DoesMediaBelongToPost(ctx context.Context, postID, mediaID string) (bool, error)
	IsPlatformEnabledForProject(ctx context.Context, projectID, platformID string) (bool, error)
	IsThePostEnabledToPlatform(ctx context.Context, postID, platformID string) (bool, error)
	IsMediaLinkedToPublishPost(ctx context.Context, postID, mediaID, platformID string) (bool, error)
	ListMediaFilesForPost(ctx context.Context, postID string) ([]string, error)
	GetMediaFileName(ctx context.Context, mediaID string) (string, error)
	DeleteMetadata(ctx context.Context, mediaID string) error
}
