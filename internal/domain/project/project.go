package project

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProjectNotFound              = errors.New("project not found")
	ErrNoUserIDInContext            = errors.New("no user id in context")
	ErrUserAlreadyInProject         = errors.New("user is already in project")
	ErrUserNotInProject             = errors.New("user is not in project")
	ErrUserNotFound                 = errors.New("user not found")
	ErrProjectExists                = errors.New("project already exists")
	ErrSocialPlatformNotFound       = errors.New("social network not found")
	ErrSocialPlatformAlreadyEnabled = errors.New("social network already enabled")
	ErrSocialPlatformNotEnabled     = errors.New("social network not enabled")
)

type TeamRoleOptions string

type SocialPlatform struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

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
	IdeaQueue   []string  `json:"idea_queue"`
	PostQueue   []string  `json:"post_queue"`
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
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	DefaultUser bool      `json:"default_user"`
	AddedAt     time.Time `json:"added_at"`
	MaxRole     int       `json:"max_role"`
}

type UserPlatformInfo struct {
	IsAuthenticated bool
	AuthTTL         time.Time
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
		PostQueue:   []string{},
		IdeaQueue:   []string{},
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}
