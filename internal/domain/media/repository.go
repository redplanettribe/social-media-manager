package media

import "context"

type Repository interface {
	SaveMetadata(ctx context.Context, media *MetaData) (*MetaData, error)
	GetMetadata(ctx context.Context, postID, fileName string) (*MetaData, error)
}