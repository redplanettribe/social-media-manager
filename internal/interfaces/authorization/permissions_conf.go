package authorization

func GetPermissions() *Permissions {
	return NewPermissions().
		AddRole("base").
		Read("users").
		Write("users").
		AddRole("manager").Inherit("base").
		Read("roles")
}
