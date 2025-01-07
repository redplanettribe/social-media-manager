package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
)

type MediaRepository struct {
	db *pgxpool.Pool
}

func NewMediaRepository(db *pgxpool.Pool) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) SaveMetadata(ctx context.Context, m *media.MetaData) (*media.MetaData, error) {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (id, post_id, media_type, media_url, thumbnail_url, width, height, length, added_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, Media), m.ID, m.PostID, m.Type, m.Url, m.ThumbnailUrl, m.Width, m.Height, m.Length, m.AddedBy, m.CreatedAt)
	if err != nil {
		return nil, err
	}
	
	return m, nil
}