package postgres

type TableNames string

const (
	Projects         TableNames = "projects"
	Users            TableNames = "users"
	TeamMembers      TableNames = "team_members"
	TeamMembersRoles TableNames = "team_members_roles"
	TeamRoles        TableNames = "team_roles"
	Posts            TableNames = "posts"
	Platforms        TableNames = "platforms"
	PostPlatform     TableNames = "post_platforms"
	ProjectPlatform  TableNames = "project_platforms"
)
