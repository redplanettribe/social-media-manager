package authorization

func GetTeamPermissions() *Permissions {
	return NewPermissions().
		AddRole("member").
		/* */ Read("projects").
		/* */ Read("posts").
		/* */ Write("posts").
		/* */ Delete("posts").
		AddRole("manager").Inherit("member").
		/* */ Write("projects").
		AddRole("owner").Inherit("manager")
}
