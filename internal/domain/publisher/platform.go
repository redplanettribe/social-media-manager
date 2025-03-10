package publisher

import "errors"

var (
	ErrSocialPlatformNotFound             = errors.New("social network not found")
	ErrSocialPlatformNotEnabledForProject = errors.New("social network not enabled for project")
	ErrDefaultUserNotSet                  = errors.New("default user not set")
	ErrNoPublishersAssigned               = errors.New("no publishers assigned")
	ErrPlatformSecretsNotSet              = errors.New("platform secrets not set")
	ErrUserSecretsNotSet                  = errors.New("user secrets not set")
	ErrOneFilePerVideoPost                = errors.New("only one file per video post")
)

// up to 10 characters
type PlatformID string

const (
	Facebook  PlatformID = "facebook"
	X         PlatformID = "x"
	LinkedIn  PlatformID = "linkedin"
	Instagram PlatformID = "instagram"
	// ...
)

type Platform struct {
	ID   PlatformID
	Name string
}
