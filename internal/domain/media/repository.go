package media

import "context"

type Repository interface {
	SaveMetadata(ctx context.Context, media *MetaData) (*MetaData, error)
}