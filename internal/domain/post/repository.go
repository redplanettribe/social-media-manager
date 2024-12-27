package post

import (
	"context"
)

type Repository interface {
	Save(ctx context.Context, post *Post) error
	FindByID(ctx context.Context, id string) (*Post, error)
	FindByProjectID(ctx context.Context, projecID string) ([]*Post, error)
	ArchivePost(ctx context.Context, id string) error
	DeletePost(ctx context.Context, id string) error
	AddSocialMediaPublisher(ctx context.Context, postID, publisherID string) error
	FindScheduledReadyPosts(ctx context.Context, offset, chunksize int) ([]*QPost, error)
}
