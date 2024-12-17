package authorization

func GetTeamPermissions() *Permissions {
	return NewPermissions().
		AddRole("member").
		/* */ Read("projects").
		AddRole("manager").Inherit("member").
		AddRole("owner").Inherit("manager")
}
