package media

import "context"

type ObjectRepository interface {
	UploadFile(ctx context.Context, projectID, postID, filename string, data []byte, metadata *MetaData) error
	GetSignedURL(ctx context.Context, projectID, postID, fileName string) (string, error)
	GetFile(ctx context.Context, projectID, postID, filename string) ([]byte, error)
	DeleteFile(ctx context.Context, projectID, postID, filename string) error
}
