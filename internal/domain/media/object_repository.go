package media

import "context"

type ObjectRepository interface {
	UploadFile(ctx context.Context, projectID, postID, filename string, data []byte) error
}