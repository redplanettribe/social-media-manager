package postgres

type TableNames string

const (
	Projects           TableNames = "projects"
	Users              TableNames = "users"
	TeamMembers        TableNames = "team_members"
	TeamMembersRoles   TableNames = "team_members_roles"
	TeamRoles          TableNames = "team_roles"
	Posts              TableNames = "posts"
	SocialNetworks     TableNames = "social_networks"
	PostSocialNetworks TableNames = "post_social_networks"
)
