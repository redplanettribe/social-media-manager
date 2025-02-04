package authorization

func GetTeamPermissions() *Permissions {
	return NewPermissions().
		AddRole("member").
		/* */ Read("projects").
		/* */ Delete("projects").
		/* */ Read("posts").
		/* */ Write("posts").
		/* */ Write("publishers").
		/* */ Read("publishers").
		/* */ Write("media").
		/* */ Read("media").
		/* */ Delete("media").
		AddRole("manager").Inherit("member").
		/* */ Write("projects").
		/* */ Delete("posts").
		AddRole("owner").Inherit("manager")
}
