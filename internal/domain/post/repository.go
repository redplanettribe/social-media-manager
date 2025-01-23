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
	GetSocialMediaPublishersIDs(ctx context.Context, postID string) ([]string, error)
	GetSocialMediaPlatforms(ctx context.Context, postID string) ([]Platform, error)
	FindScheduledReadyPosts(ctx context.Context, offset, chunksize int) ([]*PublishPost, error)
	SchedulePost(ctx context.Context, id string, sheduled_at time.Time) error
	IsPublisherPlatformEnabledForProject(ctx context.Context, projectID, publisherID string) (bool, error)
	GetProjectPostQueue(ctx context.Context, projectID string) (*Queue, error)
	AddToProjectQueue(ctx context.Context, projectID, postID string) error
	GetProjectQueuedPosts(ctx context.Context, projectID string, postIDs []string) ([]*Post, error)
	UpdateProjectPostQueue(ctx context.Context, projectID string, queue []string) error
	GetPostsForPublishQueue(ctx context.Context, postID string) ([]*PublishPost, error)
	GetPostToPublish(ctx context.Context, id string) (*PublishPost, error)
}
