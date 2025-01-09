package media

import "context"

type Repository interface {
	SaveMetadata(ctx context.Context, media *MetaData) (*MetaData, error)
	GetMetadata(ctx context.Context, postID, fileName string) (*MetaData, error)
	GetMediaNamesForPost(ctx context.Context, projectID, postID, platformID string) ([]string, error)
	LinkMediaToPublishPost(ctx context.Context, postID, fileName, platformID string) error
	DoesPostBelongToProject(ctx context.Context, projectID, postID string) (bool, error)
	DoesMediaBelongToPost(ctx context.Context, postID, mediaID string) (bool, error)
	IsPlatformEnabledForProject(ctx context.Context, projectID, platformID string) (bool, error)
	IsThePostEnabledToPlatform(ctx context.Context, postID, platformID string) (bool, error)
}
