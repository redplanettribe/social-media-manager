package media

import "context"

type ObjectRepository interface {
	// UploadFile uploads a file to the object storage, assigns the url to the metadata according if it is a thumbnail or not (to be a thumbnail, the file name must start with "th-")
	UploadFile(ctx context.Context, projectID, postID, filename string, data []byte, metadata *MetaData) error
	GetFile(ctx context.Context, projectID, postID, filename string) ([]byte, error)
}
