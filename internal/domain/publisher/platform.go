package publisher

import "errors"

var (
	ErrSocialPlatformNotFound             = errors.New("social network not found")
	ErrSocialPlatformNotEnabledForProject = errors.New("social network not enabled for project")
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
