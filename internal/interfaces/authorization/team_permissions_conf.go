package authorization

func GetTeamPermissions() *Permissions {
	return NewPermissions().
		AddRole("member").
		/* */ Read("projects").
		AddRole("manager").Inherit("member").
		/* */ Write("projects").
		AddRole("owner").Inherit("manager")
}
