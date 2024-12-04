package authorization

import "testing"

func Test_HasPermission(t *testing.T) {
	t.Run("should return true if the role has the permission", func(t *testing.T) {
		p := NewPermissions().AddRole("admin").addPermissions("read", "users")
		if !p.HasPermission(NewRoles([]string{"admin"}), "read", "users") {
			t.Errorf("expected role admin to have permission read on users, got %v", p)
		}
	})
	t.Run("should return false if the role does not have the permission", func(t *testing.T) {
		p := NewPermissions().AddRole("admin").addPermissions("read", "users")
		if p.HasPermission(NewRoles([]string{"admin"}), "write", "users") {
			t.Errorf("expected role admin to not have permission write on users, got %v", p)
		}
	})
	t.Run("should return false if the role does not have the resource", func(t *testing.T) {
		p := NewPermissions().AddRole("admin").addPermissions("read", "users")
		if p.HasPermission(NewRoles([]string{"admin"}), "read", "posts") {
			t.Errorf("expected role admin to not have permission read on posts, got %v", p)
		}
	})
	t.Run("should return false if the role does not have the action", func(t *testing.T) {
		p := NewPermissions().AddRole("admin").addPermissions("read", "users")
		if p.HasPermission(NewRoles([]string{"admin"}), "write", "users") {
			t.Errorf("expected role admin to not have permission write on users, got %v", p)
		}
	})
	t.Run("should return false if the role does not exist", func(t *testing.T) {
		p := NewPermissions().AddRole("admin").addPermissions("read", "users")
		if p.HasPermission(NewRoles([]string{"user"}), "read", "users") {
			t.Errorf("expected role user to not have permission read on users, got %v", p)
		}
	})
}

func Test_AddPermission(t *testing.T) {
	t.Run("should add a new permission", func(t *testing.T) {
		p := NewPermissions().AddRole("admin").addPermissions("read", "users")
		if _, exists := p.roles["admin"]["read"]; !exists {
			t.Errorf("expected permission read to be added, got %v", p)
		}
	})
	t.Run("should add multiple resources to a permission", func(t *testing.T) {
		p := NewPermissions().AddRole("admin").addPermissions("read", "users", "posts")
		if len(p.roles["admin"]["read"]) != 2 {
			t.Errorf("expected two resources to be added to permission read, got %v", p)
		}
	})
	t.Run("should not add a permission to a non-existing role", func(t *testing.T) {
		p := NewPermissions().addPermissions("read", "users")
		if _, exists := p.roles[""]["read"]; exists {
			t.Errorf("expected permission read to not be added, got %v", p)
		}
	})
	t.Run("should not add an existing permission", func(t *testing.T) {
		p := NewPermissions().AddRole("admin").addPermissions("read", "users").addPermissions("read", "users")
		if len(p.roles["admin"]["read"]) > 1 {
			t.Errorf("expected only one permission read to be added, got %v", p)
		}
	})
}

func Test_AddRole(t *testing.T) {
	t.Run("should add a new role", func(t *testing.T) {
		p := NewPermissions().AddRole("admin")
		if _, exists := p.roles["admin"]; !exists {
			t.Errorf("expected role admin to be added, got %v", p)
		}
	})
	t.Run("should not add an existing role", func(t *testing.T) {
		p := NewPermissions().AddRole("admin").AddRole("admin")
		if len(p.roles) > 1 {
			t.Errorf("expected only one role admin to be added, got %v", p)
		}
	})
	t.Run("current role should be set to the last added role", func(t *testing.T) {
		p := NewPermissions().AddRole("user").AddRole("admin")
		if p.currentRole != "admin" {
			t.Errorf("expected current role to be admin, got %v", p.currentRole)
		}
	})
}
