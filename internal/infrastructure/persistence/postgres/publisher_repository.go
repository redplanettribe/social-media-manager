package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redplanettribe/social-media-manager/internal/domain/publisher"
)

type PublisherRepository struct {
	db *pgxpool.Pool
}

func NewPublisherRepository(db *pgxpool.Pool) *PublisherRepository {
	return &PublisherRepository{db: db}
}

func (r *PublisherRepository) FindAll(ctx context.Context) ([]publisher.Platform, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, name
		FROM %s
	`, Platforms))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sns []publisher.Platform
	for rows.Next() {
		sn := publisher.Platform{}
		err := rows.Scan(&sn.ID, &sn.Name)
		if err != nil {
			return nil, err
		}
		sns = append(sns, sn)
	}
	return sns, nil
}

func (r *PublisherRepository) FindByID(ctx context.Context, id string) (*publisher.Platform, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(
		`SELECT id, name FROM %s WHERE id = $1`, Platforms), id)

	sn := &publisher.Platform{}
	err := row.Scan(&sn.ID, &sn.Name)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return sn, nil
}

// Checks if a record on the ProjectPlatform table exists for the given project and social publisher. It should be only one row
func (r *PublisherRepository) IsSocialNetworkEnabledForProject(ctx context.Context, projectID, socialPlatformID string) (bool, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s
		WHERE project_id = $1 AND platform_id = $2
	`, ProjectPlatforms), projectID, socialPlatformID)

	var count int
	err := row.Scan(&count)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}

	return count > 0, nil
}

func (r *PublisherRepository) GetPlatformSecrets(ctx context.Context, projectID, socialPlatformID string) (*string, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT secrets
		FROM %s
		WHERE project_id = $1 AND platform_id = $2
	`, ProjectPlatforms), projectID, socialPlatformID)

	var secrets *string
	err := row.Scan(&secrets)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return secrets, nil
}

func (r *PublisherRepository) SetPlatformSecrets(ctx context.Context, projectID, socialPlatformID, secrets string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET secrets = $3
		WHERE project_id = $1 AND platform_id = $2
	`, ProjectPlatforms), projectID, socialPlatformID, secrets)
	if err != nil {
		return err
	}

	return nil
}

func (r *PublisherRepository) GetUserPlatformSecrets(ctx context.Context, platformID, userID string) (string, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT secrets
		FROM %s
		WHERE platform_id = $1 AND user_id = $2
	`, UserPlatforms), platformID, userID)

	var secrets string
	err := row.Scan(&secrets)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	return secrets, nil
}

func (r *PublisherRepository) SetUserPlatformSecrets(ctx context.Context, platformID, userID, secrets string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (platform_id, user_id, secrets)
		VALUES ($1, $2, $3)
		ON CONFLICT (platform_id, user_id)
		DO UPDATE SET secrets = $3
	`, UserPlatforms), platformID, userID, secrets)
	if err != nil {
		return err
	}
	return nil
}

func (r *PublisherRepository) GetDefaultUserID(ctx context.Context, projectID string) (string, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT user_id
		FROM %s
		WHERE project_id = $1 AND default_user = true
	`, TeamMembers), projectID)

	var userID string
	err := row.Scan(&userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}

	return userID, nil
}

func (r *PublisherRepository) SetUserPlatformAuthSecretsWithTTL(ctx context.Context, platformID, userID, secrets string, ttl time.Time) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (platform_id, user_id, secrets, is_authenticated, auth_ttl)
		VALUES ($1, $2, $3, true, $4)
		ON CONFLICT (platform_id, user_id)
		DO UPDATE SET secrets = $3, is_authenticated = true, auth_ttl = $4
	`, UserPlatforms), platformID, userID, secrets, ttl)
	if err != nil {
		return err
	}
	return nil
}
