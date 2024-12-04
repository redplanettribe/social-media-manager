package authorization

// We could also implement ReadOwn, WriteOwn, DeleteOwn permissions.
// For now we will keep it simple. And let the user implement it if needed.

import (
	"fmt"
)

type RolePermissions map[string]map[string]map[string]string

type Permissions struct {
	roles       RolePermissions
	currentRole string
}

func NewPermissions() *Permissions {
	return &Permissions{
		roles:       make(RolePermissions),
		currentRole: "",
	}
}

type Roles map[string]struct{}

func NewRoles(roles []string) *Roles {
	r := make(Roles)
	for _, role := range roles {
		r[role] = struct{}{}
	}
	return &r
}

func (p *Permissions) AddRole(role string) *Permissions {
	if _, exists := p.roles[role]; !exists {
		p.roles[role] = make(map[string]map[string]string)
	}
	p.currentRole = role
	return p
}

func (p *Permissions) addPermissions(action string, resources ...string) *Permissions {
	if p.currentRole == "" {
		fmt.Printf("Role %s does not exist, ignoring %s action \n", p.currentRole, action)
		return p
	}
	if _, exists := p.roles[p.currentRole]; !exists {
		p.AddRole(p.currentRole)
	}
	if _, exists := p.roles[p.currentRole][action]; !exists {
		p.roles[p.currentRole][action] = make(map[string]string, 0)
	}

	for _, resource := range resources {
		if _, exists := p.roles[p.currentRole][action][resource]; !exists {
			p.roles[p.currentRole][action][resource] = resource
		}
	}
	return p
}

func (p *Permissions) Write(resources ...string) *Permissions {
	return p.addPermissions("write", resources...)
}

func (p *Permissions) Read(resources ...string) *Permissions {
	return p.addPermissions("read", resources...)
}

func (p *Permissions) Delete(resources ...string) *Permissions {
	return p.addPermissions("delete", resources...)
}

// InheritRole adds the permissions of an existing role to the current role.
func (p *Permissions) Inherit(parent string) *Permissions {
	if _, exists := p.roles[parent]; !exists {
		fmt.Printf("Role %s does not exist\n", parent)
		return p
	}
	if _, exists := p.roles[p.currentRole]; !exists {
		p.AddRole(p.currentRole)
	}
	for action, resources := range p.roles[parent] {
		if _, exists := p.roles[p.currentRole][action]; !exists {
			p.roles[p.currentRole][action] = make(map[string]string, 0)
		}
		for resource := range resources {
			if _, exists := p.roles[p.currentRole][action][resource]; !exists {
				p.roles[p.currentRole][action][resource] = resource
			}
		}
	}
	return p
}

func (p *Permissions) HasPermission(roles *Roles, action, resource string) bool {
	for role := range *roles {
		if _, exists := p.roles[role]; exists {
			if _, exists := p.roles[role][action]; exists {
				if _, exists := p.roles[role][action][resource]; exists {
					return true
				}
			}
		}
	}
	return false
}
