package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pedrodcsjostrom/opencm/internal/domain/platform"
)

type PlatformRepository struct {
	db *pgxpool.Pool
}

func NewPlatformRepository(db *pgxpool.Pool) *PlatformRepository {
	return &PlatformRepository{db: db}
}

func (r *PlatformRepository) FindAll(ctx context.Context) ([]platform.Platform, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, name
		FROM %s
	`, Platforms))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sns []platform.Platform
	for rows.Next() {
		sn := platform.Platform{}
		err := rows.Scan(&sn.ID, &sn.Name)
		if err != nil {
			return nil, err
		}
		sns = append(sns, sn)
	}
	return sns, nil
}

func (r *PlatformRepository) FindByID(ctx context.Context, id string) (*platform.Platform, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(
		`SELECT id, name FROM %s WHERE id = $1`, Platforms), id)

	sn := &platform.Platform{}
	err := row.Scan(&sn.ID, &sn.Name)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return sn, nil
}

// Checks if a record on the ProjectPlatform table exists for the given project and social platform. It should be only one row
func (r *PlatformRepository) IsSocialNetworkEnabledForProject(ctx context.Context, projectID, socialPlatformID string) (bool, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE project_id = $1 AND platform_id = $2
	`, ProjectPlatform), projectID, socialPlatformID)

	var count int
	err := row.Scan(&count)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}

	return count > 0, nil
}

func (r *PlatformRepository) AddAPIKey(ctx context.Context, projectID, socialPlatformID, apiKey string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET api_key = $1
		WHERE project_id = $2 AND platform_id = $3
	`, ProjectPlatform), apiKey, projectID, socialPlatformID)

	return err
}
