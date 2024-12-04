package authorization

func GetPermissions() *Permissions {
	return NewPermissions().
		AddRole("base").
		AddPermissions("read", "users").
		AddPermissions("write", "users").
		AddRole("manager").Inherit("base").
		AddPermissions("read", "roles")
}
