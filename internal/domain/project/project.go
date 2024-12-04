package project

import (
	"errors"
	"time"

	"github.com/google/uuid"
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
	ID          string
	Name        string
	Description string
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TeamRole struct {
	ID   string
	Name string
}

type TeamMember struct {
	ID      string
	AddedAt string
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
