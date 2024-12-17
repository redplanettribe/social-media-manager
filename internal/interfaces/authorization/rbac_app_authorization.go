package authorization

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrEmptyPermission  = errors.New("permission string is empty")
	ErrFailedToGetRoles = errors.New("failed to get user roles")
	ErrPermissionDenied = errors.New("permission denied")
)

// AppAuthorizer is the interface that wraps the Authorize method.
type AppAuthorizer interface {
	Authorize(ctx context.Context, userID string, permission string) error
}


// RBACAppAuthorizer uses Role-Based Access Control (RBAC) to determine if a user has a specific permission.
type RBACAppAuthorizer struct {
	permissions *Permissions
	getRoles    func(ctx context.Context, userID string) ([]string, error)
}

// NewAppAuthorizer creates a new RBACAuthorizer with the given role-permission and user-role mappings.
func NewAppAuthorizer(permissions *Permissions, getRoles func(ctx context.Context, userID string) ([]string, error)) AppAuthorizer {
	return &RBACAppAuthorizer{
		permissions: permissions,
		getRoles:    getRoles,
	}
}

// Authorize checks if the user with the given ID has the specified permission.
// It returns true if the user has the permission, false otherwise.
// The permission string should be in the format "action:resource".
func (a *RBACAppAuthorizer) Authorize(ctx context.Context, userID string, permission string) error {
	action, resource := parsePermission(permission)
	if action == "" || resource == "" {
		return ErrEmptyPermission
	}
	userRoles, err := a.getRoles(ctx, userID)
	if err != nil {
		return errors.Join(ErrFailedToGetRoles, err)
	}
	roleMap := NewRoles(userRoles)
	hasPermission := a.permissions.HasPermission(roleMap, action, resource)
	if !hasPermission {
		return ErrPermissionDenied
	}
	return nil
}

// parsePermission parses the given permission string into an action and a resource.
// The permission string should be in the format "action:resource".
// It returns the action and resource strings.
func parsePermission(permission string) (action, resource string) {
	parts := strings.Split(permission, ":")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}