package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pedrodcsjostrom/opencm/internal/domain/project"
)

type ProjectRepository struct {
	db *pgx.Conn
}

type TableNames string

const (
	Projects         TableNames = "projects"
	TeamMembers      TableNames = "team_members"
	TeamMembersRoles TableNames = "team_members_roles"
	TeamRoles        TableNames = "team_roles"
)

func NewProjectRepository(db *pgx.Conn) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Save(ctx context.Context, p *project.Project) (*project.Project, error) {
	_, err := r.db.Exec(ctx, fmt.Sprintf(`
		INSERT INTO %s (id, name, description, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, Projects), p.ID, p.Name, p.Description, p.CreatedBy, time.Now(), time.Now())
	if err != nil {
		return &project.Project{}, err
	}

	return p, nil
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

	_, err = tx.Exec(ctx, insertMember, projectID, userID, time.Now())
	if err != nil {
		return err
	}

	// Prepare the INSERT statement for team_members_roles
	insertRoleStmt := fmt.Sprintf(`
        INSERT INTO %s (project_id,team_role_id, user_id)
        VALUES ($1, $2, $3)
    `, TeamMembersRoles)

	// Batch insert roleIDs
	roleIDs := []project.TeamRoleIDs{project.OwnerRoleID, project.ManagerRoleID, project.MemberRoleID}
	for _, roleID := range roleIDs {
		_, err = tx.Exec(ctx, insertRoleStmt, projectID, roleID, userID)
		if err != nil {
			return err
		}
	}

	return nil
}
