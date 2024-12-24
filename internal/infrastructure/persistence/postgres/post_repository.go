package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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
	// Implementation...
	return nil, nil
}

func (r *PostRepository) FindByProjectID(ctx context.Context, teamID uuid.UUID) ([]*post.Post, error) {
	// Implementation...
	return nil, nil
}
