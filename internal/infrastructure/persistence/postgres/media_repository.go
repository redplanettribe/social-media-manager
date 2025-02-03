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
            id, post_id, file_name, media_type, format, width, height, length, size, alt_text, added_by, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `, Media),
		m.ID, m.PostID, m.Filename, m.Type, m.Format, m.Width, m.Height, m.Length, m.Size, m.AltText, m.AddedBy, m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *MediaRepository) GetMetadata(ctx context.Context, postID, fileName string) (*media.MetaData, error) {
	var m media.MetaData
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, post_id, file_name, media_type, format, width, height, length, size, alt_text, added_by, created_at
		FROM %s
		WHERE post_id = $1 AND file_name = $2
	`, Media), postID, fileName).Scan(
		&m.ID, &m.PostID, &m.Filename, &m.Type, &m.Format, &m.Width, &m.Height, &m.Length, &m.Size, &m.AltText, &m.AddedBy, &m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MediaRepository) GetMediaFileNamesForPublishPost(ctx context.Context, postID, platformID string) ([]string, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT file_name
		FROM %s m
		JOIN %s ppm ON m.id = ppm.media_id
		WHERE ppm.post_id = $1 AND ppm.platform_id = $2
	`, Media, PostPlatformMedia), postID, platformID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fileNames []string
	for rows.Next() {
		var fileName string
		err := rows.Scan(&fileName)
		if err != nil {
			return nil, err
		}
		fileNames = append(fileNames, fileName)
	}
	return fileNames, nil
}

func (r *MediaRepository) LinkMediaToPublishPost(ctx context.Context, postID, mediaID, platformID string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (post_id, media_id, platform_id)
		VALUES ($1, $2, $3)
	`, PostPlatformMedia), postID, mediaID, platformID)
	if err != nil {
		return err
	}
	return nil
}

func (r *MediaRepository) UnlinkMediaFromPublishPost(ctx context.Context, postID, mediaID, platformID string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE post_id = $1 AND media_id = $2 AND platform_id = $3
	`, PostPlatformMedia), postID, mediaID, platformID)
	if err != nil {
		return err
	}
	return nil
}

func (r *MediaRepository) DoesPostBelongToProject(ctx context.Context, projectID, postID string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE id = $1 AND project_id = $2
	`, Posts), postID, projectID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MediaRepository) DoesMediaBelongToPost(ctx context.Context, postID, mediaID string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE post_id = $1 AND id = $2
	`, Media), postID, mediaID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MediaRepository) IsPlatformEnabledForProject(ctx context.Context, projectID, platformID string) (bool, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE project_id = $1 AND platform_id = $2
	`, ProjectPlatforms), projectID, platformID)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MediaRepository) IsThePostEnabledToPlatform(ctx context.Context, postID, platformID string) (bool, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE post_id = $1 AND platform_id = $2
	`, PostPlatforms), postID, platformID)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MediaRepository) IsMediaLinkedToPublishPost(ctx context.Context, postID, mediaID, platformID string) (bool, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE post_id = $1 AND media_id = $2 AND platform_id = $3
	`, PostPlatformMedia), postID, mediaID, platformID)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MediaRepository) ListMediaFilesForPost(ctx context.Context, postID string) ([]string, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT file_name
		FROM %s
		WHERE post_id = $1 AND file_name NOT LIKE 'thumbnail_%%'
	`, Media), postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fileNames []string
	for rows.Next() {
		var fileName string
		err := rows.Scan(&fileName)
		if err != nil {
			return nil, err
		}
		fileNames = append(fileNames, fileName)
	}
	return fileNames, nil
}
