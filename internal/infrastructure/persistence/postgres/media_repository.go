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
        INSERT INTO %s (
            id, post_id, file_name, media_type, format, width, height, length, added_by, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `, Media),
		m.ID, m.PostID, m.Filename, m.Type, m.Format, m.Width, m.Height, m.Length, m.AddedBy, m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *MediaRepository) GetMetadata(ctx context.Context, postID, fileName string) (*media.MetaData, error) {
	fmt.Println("filename", fileName)
	var m media.MetaData
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, post_id, file_name, media_type, format, width, height, length, added_by, created_at
		FROM %s
		WHERE post_id = $1 AND file_name = $2
	`, Media), postID, fileName).Scan(
		&m.ID, &m.PostID, &m.Filename, &m.Type, &m.Format, &m.Width, &m.Height, &m.Length, &m.AddedBy, &m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
