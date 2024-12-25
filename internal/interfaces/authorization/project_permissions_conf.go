package authorization

func GetTeamPermissions() *Permissions {
	return NewPermissions().
		AddRole("member").
		/* */ Read("projects").
		/* */ Read("posts").
		/* */ Write("posts").
		AddRole("manager").Inherit("member").
		/* */ Write("projects").
		/* */ Delete("posts").
		AddRole("owner").Inherit("manager")
}
