package project

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProjectNotFound      = errors.New("project not found")
	ErrNoUserIDInContext    = errors.New("no user id in context")
	ErrUserAlreadyInProject = errors.New("user is already in project")
	ErrUserNotFound         = errors.New("user not found")
	ErrProjectExists = errors.New("project already exists")
)

type TeamRoleOptions string

const (
	ManagerRole TeamRoleOptions = "manager"
	MemberRole  TeamRoleOptions = "member"
	OwnerRole   TeamRoleOptions = "owner"
)

type TeamRoleIDs int

const (
	ManagerRoleID TeamRoleIDs = 1
	MemberRoleID  TeamRoleIDs = 2
	OwnerRoleID   TeamRoleIDs = 3
)

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProjectResponse struct {
	Project *Project
	Users   []*TeamMember
}

type TeamRole struct {
	ID   string
	Name string
}

type TeamMember struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	AddedAt time.Time `json:"added_at"`
	MaxRole int       `json:"max_role"`
}

func NewProject(name, description, createdBy string) (*Project, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if description == "" {
		return nil, errors.New("description cannot be empty")
	}
	if createdBy == "" {
		return nil, errors.New("createdBy cannot be empty")
	}

	return &Project{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}
