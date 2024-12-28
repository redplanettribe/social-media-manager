package post

import (
	"context"
	time "time"
)

type Repository interface {
	Save(ctx context.Context, post *Post) error
	Update(ctx context.Context, post *Post) error
	FindByID(ctx context.Context, id string) (*Post, error)
	FindByProjectID(ctx context.Context, projecID string) ([]*Post, error)
	ArchivePost(ctx context.Context, id string) error
	DeletePost(ctx context.Context, id string) error
	AddSocialMediaPublisher(ctx context.Context, postID, publisherID string) error
	FindScheduledReadyPosts(ctx context.Context, offset, chunksize int) ([]*QPost, error)
	SchedulePost(ctx context.Context, id string, sheduled_at time.Time) error
	IsPublisherPlatformEnabledForProject(ctx context.Context, projectID, publisherID string) (bool, error)
	GetProjectPostQueue(ctx context.Context, projectID string) (*Queue, error)
	AddToProjectQueue(ctx context.Context, projectID, postID string) error
	GetProjectQueuedPosts(ctx context.Context, projectID string, postIDs []string) ([]*Post, error)
	UpdateProjectPostQueue(ctx context.Context, projectID string, queue []string) error
	GetPostsForPlatformPublishQueue(ctx context.Context, postID string) ([]*QPost, error)
}
