package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redplanettribe/social-media-manager/internal/domain/project"
)

type ProjectRepository struct {
	db *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Save(ctx context.Context, p *project.Project) (*project.Project, error) {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (id, name, description, post_queue, idea_queue, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, Projects), p.ID, p.Name, p.Description, p.PostQueue, p.IdeaQueue, p.CreatedBy, time.Now().UTC(), time.Now().UTC())
	if err != nil {
		return &project.Project{}, err
	}

	return p, nil
}

func (r *ProjectRepository) Update(ctx context.Context, p *project.Project) (*project.Project, error) {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4
		RETURNING id, name, description, post_queue, idea_queue, created_by, created_at, updated_at
	`, Projects), p.Name, p.Description, time.Now().UTC(), p.ID)
	if err != nil {
		return &project.Project{}, err
	}

	return p, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, projectID string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE id = $1
	`, Projects), projectID)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProjectRepository) ListByUserID(ctx context.Context, userID string) ([]*project.Project, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.post_queue, p.idea_queue, p.created_by, p.created_at, p.updated_at
		FROM %s p
		INNER JOIN %s tm ON p.id = tm.project_id
		WHERE tm.user_id = $1
	`, Projects, TeamMembers), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*project.Project
	for rows.Next() {
		p := &project.Project{}
		err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.PostQueue, &p.IdeaQueue, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, nil
}

func (r *ProjectRepository) AssignProjectOwner(ctx context.Context, projectID, userID string) error {
	// Begin a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	// Insert into team_members
	insertMember := fmt.Sprintf(`
        INSERT INTO %s (project_id, user_id, added_at)
        VALUES ($1, $2, $3)
    `, TeamMembers)

	_, err = tx.Exec(ctx, insertMember, projectID, userID, time.Now().UTC())
	if err != nil {
		return err
	}

	// Prepare the INSERT statement for team_members_roles
	insertRoleStmt := fmt.Sprintf(`
        INSERT INTO %s (project_id,team_role_id, user_id)
        VALUES ($1, $2, $3)
    `, TeamMembersRoles)

	// Batch insert roleIDs
	roleIDs := []project.TeamRoleID{project.OwnerRoleID, project.ManagerRoleID, project.MemberRoleID}
	for _, roleID := range roleIDs {
		_, err = tx.Exec(ctx, insertRoleStmt, projectID, roleID, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ProjectRepository) GetUserRoles(ctx context.Context, userID, projectID string) ([]string, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT tr.role
		FROM %s tmr
		INNER JOIN %s tr ON tmr.team_role_id = tr.id
		WHERE tmr.user_id = $1 AND tmr.project_id = $2
	`, TeamMembersRoles, TeamRoles), userID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var role string
		err = rows.Scan(&role)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *ProjectRepository) GetUserMaxRole(ctx context.Context, userID, projectID string) (int, error) {
	var roleID int
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT MAX(team_role_id)
		FROM %s
		WHERE user_id = $1 AND project_id = $2
	`, TeamMembersRoles), userID, projectID).Scan(&roleID)
	if err != nil {
		return 0, err
	}

	return roleID, nil
}

func (r *ProjectRepository) AddUserRole(ctx context.Context, projectID, userID string, roleID project.TeamRoleID) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
        INSERT INTO %s (project_id,team_role_id, user_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (project_id, team_role_id, user_id) DO NOTHING
    `, TeamMembersRoles), projectID, roleID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProjectRepository) RemoveUserRole(ctx context.Context, projectID, userID string, roleID project.TeamRoleID) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE project_id = $1 AND user_id = $2 AND team_role_id = $3
	`, TeamMembersRoles), projectID, userID, roleID)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProjectRepository) FindProjectByID(ctx context.Context, projectID string) (*project.Project, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT id, name, description, post_queue, idea_queue, created_by, created_at, updated_at
		FROM %s
		WHERE id = $1
	`, Projects), projectID)

	p := &project.Project{}
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.PostQueue, &p.IdeaQueue, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *ProjectRepository) GetProjectUsers(ctx context.Context, projectID string) ([]*project.TeamMember, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT u.id, u.username, u.email, tm.default_user, tm.added_at, 
			(SELECT MAX(team_role_id)
			FROM %s tmr_sub
			WHERE tmr_sub.user_id = tm.user_id 
			AND tmr_sub.project_id = tm.project_id) as max_role_id
		FROM %s tm
		INNER JOIN %s u ON tm.user_id = u.id
		WHERE tm.project_id = $1;
	`, TeamMembersRoles, TeamMembers, Users), projectID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*project.TeamMember
	for rows.Next() {
		tm := &project.TeamMember{}
		err = rows.Scan(&tm.ID, &tm.Name, &tm.Email, &tm.DefaultUser, &tm.AddedAt, &tm.MaxRole)
		if err != nil {
			return nil, err
		}
		users = append(users, tm)
	}

	return users, nil
}

func (r *ProjectRepository) IsUserInProject(ctx context.Context, projectID, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1
			FROM %s
			WHERE project_id = $1 AND user_id = $2
		)
	`, TeamMembers), projectID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *ProjectRepository) AddUserToProject(ctx context.Context, projectID, userID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()
	// Insert into team_members
	_, err = tx.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (project_id, user_id, added_at)
		VALUES ($1, $2, $3)
	`, TeamMembers), projectID, userID, time.Now().UTC())
	if err != nil {
		return err
	}
	// Insert into team_members_roles
	_, err = tx.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (project_id, team_role_id, user_id)
		VALUES ($1, $2, $3)
	`, TeamMembersRoles), projectID, project.MemberRoleID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProjectRepository) RemoveUserFromProject(ctx context.Context, projectID, userID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	_, err = tx.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE project_id = $1 AND user_id = $2
	`, TeamMembers), projectID, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE project_id = $1 AND user_id = $2
	`, TeamMembersRoles), projectID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProjectRepository) DoesProjectNameExist(ctx context.Context, name, userID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1
			FROM %s
			WHERE name = $1 AND created_by = $2
		)
	`, Projects), name, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *ProjectRepository) EnableSocialPlatform(ctx context.Context, projectID, socialPlatformID string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (project_id, platform_id)
		VALUES ($1, $2)
	`, ProjectPlatforms), projectID, socialPlatformID)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProjectRepository) DisableSocialPlatform(ctx context.Context, projectID, socialPlatformID string) error {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE project_id = $1 AND platform_id = $2
	`, ProjectPlatforms), projectID, socialPlatformID)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProjectRepository) DoesSocialPlatformExist(ctx context.Context, socialPlatformID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1
			FROM %s
			WHERE id = $1
		)
	`, Platforms), socialPlatformID).Scan(&exists)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	return exists, nil
}

func (r *ProjectRepository) IsProjectSocialPlatformEnabled(ctx context.Context, projectID, socialPlatformID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT EXISTS(
			SELECT 1
			FROM %s
			WHERE project_id = $1 AND platform_id = $2
		)
	`, ProjectPlatforms), projectID, socialPlatformID).Scan(&exists)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return exists, nil
}

func (r *ProjectRepository) GetEnabledSocialPlatforms(ctx context.Context, projectID string) ([]project.SocialPlatform, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT p.id, p.name
		FROM %s pp
		INNER JOIN %s p ON pp.platform_id = p.id
		WHERE pp.project_id = $1
	`, ProjectPlatforms, Platforms), projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sns := []project.SocialPlatform{}
	for rows.Next() {
		sn := project.SocialPlatform{}
		err = rows.Scan(&sn.ID, &sn.Name)
		if err != nil {
			return nil, err
		}
		sns = append(sns, sn)
	}

	return sns, nil
}

func (r *ProjectRepository) DeleteProject(ctx context.Context, projectID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	_, err = tx.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE project_id = $1
	`, ProjectPlatforms), projectID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE project_id = $1
	`, TeamMembersRoles), projectID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE project_id = $1
	`, TeamMembers), projectID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
		DELETE FROM %s
		WHERE id = $1
	`, Projects), projectID)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProjectRepository) GetProjectSchedule(ctx context.Context, projectID string) (*project.WeeklyPostSchedule, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT schedule
		FROM %s
		WHERE project_id = $1
	`, ProjectSettings), projectID)

	var encoded string
	err := row.Scan(&encoded)
	if err != nil {
		return nil, err
	}

	schedule, err := project.DecodeSchedule(encoded)
	if err != nil {
		return nil, err
	}

	return schedule, nil
}

func (r *ProjectRepository) SaveSchedule(ctx context.Context, projectID string, schedule *project.WeeklyPostSchedule) error {
	encoded, err := schedule.Encode()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET schedule = $1
		WHERE project_id = $2
	`, ProjectSettings), encoded, projectID)

	return err
}

func (r *ProjectRepository) CreateProjectSettings(ctx context.Context, projectID string, schedule *project.WeeklyPostSchedule) error {
	encoded, err := schedule.Encode()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (project_id, schedule)
		VALUES ($1, $2)
	`, ProjectSettings), projectID, encoded)

	return err
}

func (r *ProjectRepository) FindActiveProjectsChunk(ctx context.Context, offset, chunkSize int) ([]*project.Project, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, name, description, post_queue, idea_queue, created_by, created_at, updated_at
		FROM %s
		WHERE post_queue IS NOT NULL
		AND cardinality(post_queue) > 0
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, Projects), chunkSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*project.Project
	for rows.Next() {
		p := &project.Project{}
		err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.PostQueue, &p.IdeaQueue, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, nil
}

func (r *ProjectRepository) SetDefaultUser(ctx context.Context, projectID, userID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	_, err = tx.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET default_user = false
		WHERE project_id = $1
	`, TeamMembers), projectID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET default_user = true
		WHERE project_id = $1 AND user_id = $2
	`, TeamMembers), projectID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *ProjectRepository) GetDefaultUserID(ctx context.Context, projectID string) (string, error) {
	var userID string
	err := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT user_id
		FROM %s
		WHERE project_id = $1 AND default_user = true
	`, TeamMembers), projectID).Scan(&userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}

	return userID, nil
}

func (r *ProjectRepository) GetPlatformInfo(ctx context.Context, userID, platformID string) (*project.UserPlatformInfo, error) {
	row := r.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT is_authenticated, auth_ttl
		FROM %s
		WHERE user_id = $1 AND platform_id = $2
	`, UserPlatforms), userID, platformID)

	pInfo := &project.UserPlatformInfo{}
	err := row.Scan(&pInfo.IsAuthenticated, &pInfo.AuthTTL)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return pInfo, nil
}
