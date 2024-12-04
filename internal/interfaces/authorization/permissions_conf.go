package authorization

func GetPermissions() *Permissions {
	return NewPermissions().
		AddRole("user").
		Read("users").
		Write("users").
		AddRole("admin").Inherit("user").
		Read("roles").
		Write("roles").
		Delete("roles")
}
