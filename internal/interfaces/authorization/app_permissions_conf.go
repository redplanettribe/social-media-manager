package authorization

func GetAppPermissions() *Permissions {
	return NewPermissions().
		AddRole("user").
		/* */ Read("users").
		/* */ Write("users").
		/* */ Write("projects").
		/* */ Read("projects").
		/* */ Delete("projects").
		/* */ Read("publishers").
		/* */ Read("posts").
		AddRole("admin").Inherit("user").
		/* */ Read("roles").
		/* */ Write("roles").
		/* */ Delete("roles")
}
