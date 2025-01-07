package postgres

type TableNames string

const (
	Projects          TableNames = "projects"
	Users             TableNames = "users"
	TeamMembers       TableNames = "team_members"
	TeamMembersRoles  TableNames = "team_members_roles"
	TeamRoles         TableNames = "team_roles"
	Posts             TableNames = "posts"
	Media             TableNames = "media"
	Platforms         TableNames = "platforms"
	PostPlatforms     TableNames = "post_platforms"
	PostPlatformMedia TableNames = "post_platform_media"
	ProjectPlatforms  TableNames = "project_platforms"
	Comments          TableNames = "comments"
	ProjectSettings   TableNames = "project_settings"
)
