package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/redplanettribe/social-media-manager/internal/domain/post"
)

type PostRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Save(ctx context.Context, p *post.Post) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (id, project_id, title, type, text_content, is_idea, status, scheduled_at, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, Posts), p.ID, p.ProjectID, p.Title, p.Type, p.TextContent, p.IsIdea, p.Status, p.ScheduledAt, p.CreatedBy, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) Update(ctx context.Context, p *post.Post) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET title = $2, type = $3, text_content = $4, is_idea = $5, status = $6, scheduled_at = $7, updated_at = $8
		WHERE id = $1
	`, Posts), p.ID, p.Title, p.Type, p.TextContent, p.IsIdea, p.Status, p.ScheduledAt, time.Now().UTC())
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) FindByID(ctx context.Context, id string) (*post.Post, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, project_id, title, type, text_content, is_idea, status, scheduled_at, created_by, created_at, updated_at
		FROM %s
		WHERE id = $1
	`, Posts), id)

	p := &post.Post{}
	err := row.Scan(&p.ID, &p.ProjectID, &p.Title, &p.Type, &p.TextContent, &p.IsIdea, &p.Status, &p.ScheduledAt, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return p, nil
}

func (r *PostRepository) FindByProjectID(ctx context.Context, projectID string) ([]*post.Post, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, project_id, title, type, text_content, is_idea, status, scheduled_at, created_by, created_at, updated_at
		FROM %s
		WHERE project_id = $1
	`, Posts), projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*post.Post
	for rows.Next() {
		p := &post.Post{}
		err = rows.Scan(&p.ID, &p.ProjectID, &p.Title, &p.Type, &p.TextContent, &p.IsIdea, &p.Status, &p.ScheduledAt, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

func (r *PostRepository) ArchivePost(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET status = $2, updated_at = $3
		WHERE id = $1
	`, Posts), id, post.PostStatusArchived, time.Now().UTC())
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) RestorePost(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET status = $2, updated_at = $3
		WHERE id = $1
	`, Posts), id, post.PostStatusDraft, time.Now().UTC())
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) DeletePost(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE id = $1
	`, Posts), id)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) AddSocialMediaPublisher(ctx context.Context, postID, publisherID string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (post_id, platform_id, status)
		VALUES ($1, $2, $3)
	`, PostPlatforms), postID, publisherID, post.PublisherPostStatusReady)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) RemoveSocialMediaPublisher(ctx context.Context, postID, publisherID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete from post_platforms
	_, err = tx.Exec(ctx, fmt.Sprintf(`
        DELETE FROM %s
        WHERE post_id = $1 AND platform_id = $2
    `, PostPlatforms), postID, publisherID)
	if err != nil {
		return fmt.Errorf("failed to delete from post_platforms: %w", err)
	}

	// Delete from post_platform_media
	_, err = tx.Exec(ctx, fmt.Sprintf(`
        DELETE FROM %s
        WHERE post_id = $1 AND platform_id = $2
    `, PostPlatformMedia), postID, publisherID)
	if err != nil {
		return fmt.Errorf("failed to delete from post_platform_media: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *PostRepository) GetSocialMediaPublishersIDs(ctx context.Context, postID string) ([]string, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT platform_id
		FROM %s
		WHERE post_id = $1
	`, PostPlatforms), postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var publishers []string
	for rows.Next() {
		var publisher string
		err = rows.Scan(&publisher)
		if err != nil {
			return nil, err
		}
		publishers = append(publishers, publisher)
	}

	return publishers, nil
}

func (r *PostRepository) GetSocialMediaPlatforms(ctx context.Context, postID string) ([]post.Platform, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT plat.id, plat.name
		FROM %s popl
		INNER JOIN %s plat ON popl.platform_id = plat.id
		WHERE post_id = $1
	`, PostPlatforms, Platforms), postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	platforms := []post.Platform{}
	for rows.Next() {
		p := post.Platform{}
		err = rows.Scan(&p.ID, &p.Name)
		if err != nil {
			return nil, err
		}
		platforms = append(platforms, p)
	}

	return platforms, nil
}

func (r *PostRepository) FindScheduledReadyPosts(ctx context.Context, offset, chunksize int) ([]*post.PublishPost, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT DISTINCT ON (p.id)
		p.id,
		p.project_id,
		p.title,
		p.type,
		p.text_content,
		p.is_idea,
		p.status,
		p.scheduled_at,
		p.created_by,
		p.created_at,
		p.updated_at,
		prpl.secrets,
		plat.id,
		popl.status publish_status
		FROM %s p
		INNER JOIN %s popl ON p.id = popl.post_id
		INNER JOIN %s plat ON popl.platform_id = plat.id
		INNER JOIN %s prpl ON plat.id = prpl.platform_id
		WHERE p.status = $1 
		AND p.scheduled_at < $2
		AND prpl.secrets IS NOT NULL
		ORDER BY p.id, p.scheduled_at
		LIMIT $3 OFFSET $4;
		`, Posts, PostPlatforms, Platforms, ProjectPlatforms),
		post.PostStatusScheduled, time.Now().Add(5*time.Minute).UTC(), chunksize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*post.PublishPost
	for rows.Next() {
		p := &post.PublishPost{
			Post: &post.Post{},
		}
		err = rows.Scan(
			&p.Post.ID,
			&p.Post.ProjectID,
			&p.Post.Title,
			&p.Post.Type,
			&p.Post.TextContent,
			&p.Post.IsIdea,
			&p.Post.Status,
			&p.Post.ScheduledAt,
			&p.Post.CreatedBy,
			&p.Post.CreatedAt,
			&p.UpdatedAt,
			&p.Secrets,
			&p.Platform,
			&p.PublishStatus,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

func (r *PostRepository) SchedulePost(ctx context.Context, id string, scheduledAt time.Time) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET scheduled_at = $2, status = $3, updated_at = $4
		WHERE id = $1
	`, Posts), id, scheduledAt, post.PostStatusScheduled, time.Now().UTC())
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) UnschedulePost(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET scheduled_at = $2, status = $3, updated_at = $4
		WHERE id = $1
	`, Posts), id, time.Time{}, post.PostStatusDraft, time.Now().UTC())
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) IsPublisherPlatformEnabledForProject(ctx context.Context, projectID, publisherID string) (bool, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE project_id = $1 AND platform_id = $2
	`, ProjectPlatforms), projectID, publisherID)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *PostRepository) GetProjectPostQueue(ctx context.Context, projectID string) (*post.Queue, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT post_queue
		FROM %s
		WHERE id = $1
	`, Projects), projectID)

	var queue *post.Queue
	err := row.Scan(&queue)
	if err != nil {
		return queue, err
	}

	return queue, nil
}

func (r *PostRepository) UpdateProjectPostQueue(ctx context.Context, projectID string, queue []string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET post_queue = $2
		WHERE id = $1
	`, Projects), projectID, queue)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) GetProjectIdeaQueue(ctx context.Context, projectID string) (*post.Queue, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT idea_queue
		FROM %s
		WHERE id = $1
	`, Projects), projectID)

	var queue *post.Queue
	err := row.Scan(&queue)
	if err != nil {
		return queue, err
	}

	return queue, nil
}

func (r *PostRepository) UpdateProjectIdeaQueue(ctx context.Context, projectID string, queue []string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET idea_queue = $2
		WHERE id = $1
	`, Projects), projectID, queue)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) AddToProjectQueue(ctx context.Context, projectID, postID string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET post_queue = array_append(post_queue, $2)
		WHERE id = $1
	`, Projects), projectID, postID)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) RemoveFromProjectQueue(ctx context.Context, projectID, postID string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET post_queue = array_remove(post_queue, $2)
		WHERE id = $1
	`, Projects), projectID, postID)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) AddToProjectIdeaQueue(ctx context.Context, projectID, postID string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET idea_queue = array_append(idea_queue, $2)
		WHERE id = $1
	`, Projects), projectID, postID)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostRepository) RemoveFromProjectIdeaQueue(ctx context.Context, projectID, postID string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET idea_queue = array_remove(idea_queue, $2)
		WHERE id = $1
	`, Projects), projectID, postID)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostRepository) GetProjectQueuedPosts(ctx context.Context, projectID string, postIDs []string) ([]*post.Post, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT 
			id, 
			project_id, 
			title, 
			text_content, 
			image_links, 
			video_links, 
			is_idea, 
			status, 
			scheduled_at, 
			created_by, 
			created_at, 
			updated_at
		FROM %s
		WHERE project_id = $1 AND id = ANY($2)
	`, Posts), projectID, postIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*post.Post
	for rows.Next() {
		p := &post.Post{}
		err = rows.Scan(
			&p.ID,
			&p.ProjectID,
			&p.Title,
			&p.TextContent,
			&p.IsIdea,
			&p.Status,
			&p.ScheduledAt,
			&p.CreatedBy,
			&p.CreatedAt,
			&p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

func (r *PostRepository) GetPostsForPublishQueue(ctx context.Context, postID string) ([]*post.PublishPost, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
        SELECT
            p.id,
            p.project_id,
            p.title,
            p.type,
            p.text_content,
            p.is_idea,
            p.status,
            p.scheduled_at,
            p.created_by,
            p.created_at,
            p.updated_at,
            prpl.secrets,
            plat.id,
            popl.status publish_status
        FROM %s p
        INNER JOIN %s popl ON p.id = popl.post_id
        INNER JOIN %s plat ON popl.platform_id = plat.id
        INNER JOIN %s prpl ON plat.id = prpl.platform_id AND prpl.project_id = p.project_id
        WHERE p.id = $1
        AND prpl.secrets IS NOT NULL
    `, Posts, PostPlatforms, Platforms, ProjectPlatforms), postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*post.PublishPost
	for rows.Next() {
		p := &post.PublishPost{
			Post: &post.Post{},
		}
		err = rows.Scan(
			&p.Post.ID,
			&p.Post.ProjectID,
			&p.Post.Title,
			&p.Post.Type,
			&p.Post.TextContent,
			&p.Post.IsIdea,
			&p.Post.Status,
			&p.Post.ScheduledAt,
			&p.Post.CreatedBy,
			&p.Post.CreatedAt,
			&p.Post.UpdatedAt,
			&p.Secrets,
			&p.Platform,
			&p.PublishStatus,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, nil
}

func (r *PostRepository) GetPostToPublish(ctx context.Context, id string) (*post.PublishPost, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT
			p.id,
			p.project_id,
			p.title,
			p.type,
			p.text_content,
			p.is_idea,
			p.status,
			p.scheduled_at,
			p.created_by,
			p.created_at,
			p.updated_at,
			prpl.secrets,
			plat.id,
			popl.status publish_status
		FROM %s p
		INNER JOIN %s popl ON p.id = popl.post_id
		INNER JOIN %s plat ON popl.platform_id = plat.id
		INNER JOIN %s prpl ON plat.id = prpl.platform_id
		WHERE p.id = $1
	`, Posts, PostPlatforms, Platforms, ProjectPlatforms), id)

	pp := &post.PublishPost{}
	p := &post.Post{}
	err := row.Scan(
		&p.ID,
		&p.ProjectID,
		&p.Title,
		&p.Type,
		&p.TextContent,
		&p.IsIdea,
		&p.Status,
		&p.ScheduledAt,
		&p.CreatedBy,
		&p.CreatedAt,
		&p.UpdatedAt,
		&pp.Secrets,
		&pp.Platform,
		&pp.PublishStatus,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	pp.Post = p
	return pp, nil
}

func (r *PostRepository) UpdatePublishPostStatus(ctx context.Context, postID, platformID, status string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET status = $3
		WHERE post_id = $1 AND platform_id = $2
	`, PostPlatforms), postID, platformID, status)
	if err != nil {
		return err
	}
	return nil
}
