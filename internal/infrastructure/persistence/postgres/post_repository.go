package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

type PostRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Save(ctx context.Context, p *post.Post) error {
	fmt.Println("repo")
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (id, project_id, title, text_content, image_links, video_links, is_idea, status, scheduled_at, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, Posts), p.ID, p.ProjectID, p.Title, p.TextContent, p.ImageLinks, p.VideoLinks, p.IsIdea, p.Status, p.ScheduledAt, p.CreatedBy, time.Now(), time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (r *PostRepository) FindByID(ctx context.Context, id string) (*post.Post, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, project_id, title, text_content, image_links, video_links, is_idea, status, scheduled_at, created_by, created_at, updated_at
		FROM %s
		WHERE id = $1
	`, Posts), id)

	p := &post.Post{}
	err := row.Scan(&p.ID, &p.ProjectID, &p.Title, &p.TextContent, &p.ImageLinks, &p.VideoLinks, &p.IsIdea, &p.Status, &p.ScheduledAt, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return p, nil
}

func (r *PostRepository) FindByProjectID(ctx context.Context, projectID string) ([]*post.Post, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, project_id, title, text_content, image_links, video_links, is_idea, status, scheduled_at, created_by, created_at, updated_at
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
		err = rows.Scan(&p.ID, &p.ProjectID, &p.Title, &p.TextContent, &p.ImageLinks, &p.VideoLinks, &p.IsIdea, &p.Status, &p.ScheduledAt, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
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
	`, Posts), id, post.PostStatusArchived, time.Now())
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

func (r *PostRepository) FindScheduledReadyPosts(ctx context.Context, offset, chunksize int) ([]*post.QPost, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT
		p.id,
		p.project_id,
		p.title,
		p.text_content,
		p.image_links,
		p.video_links,
		p.is_idea,
		p.status,
		p.scheduled_at,
		p.created_by,
		p.created_at,
		p.updated_at,
		prpl.api_key,
		plat.id,
		popl.status
		FROM %s p
		JOIN %s popl ON p.id = popl.post_id
		JOIN %s plat ON popl.platform_id = plat.id
		JOIN %s prpl ON plat.id = prpl.platform_id
		WHERE p.status = $1
		AND p.scheduled_at <= $2
		ORDER BY p.scheduled_at
		LIMIT $3 OFFSET $4
		`, Posts, PostPlatforms, Platforms, ProjectPlatforms),
		post.PostStatusScheduled, time.Now().Add(5*time.Minute), chunksize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*post.QPost
	for rows.Next() {
		p := &post.QPost{}
		err = rows.Scan(
			&p.ID,
			&p.ProjectID,
			&p.Title,
			&p.TextContent,
			&p.ImageLinks,
			&p.VideoLinks,
			&p.IsIdea,
			&p.Status,
			&p.ScheduledAt,
			&p.CreatedBy,
			&p.CreatedAt,
			&p.UpdatedAt,
			&p.ApiKey,
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
	`, Posts), id, scheduledAt, post.PostStatusScheduled, time.Now())
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
