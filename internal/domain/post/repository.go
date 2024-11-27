package post

import "github.com/google/uuid"

type Repository interface {
	Save(post *Post) error
	FindByID(id PostID) (*Post, error)
	FindByTeamID(teamID uuid.UUID) ([]*Post, error)
	// Additional methods as needed
}
