package authorization

import (
	"context"
	"errors"
)

type TeamAuthorizer struct {
	permissions *Permissions
	getRoles    func(ctx context.Context, userID, projectID string) ([]string, error)
}

type ProjectAuthorizer interface {
	Authorize(ctx context.Context, userID, projectID, permission string) error
}

func NewTeamAthorizer(permissions *Permissions, getRoles func(ctx context.Context, userID, projectID string) ([]string, error)) ProjectAuthorizer {
	return &TeamAuthorizer{
		permissions: permissions,
		getRoles:    getRoles,
	}
}

func (a *TeamAuthorizer) Authorize(ctx context.Context, userID, projectID, permission string) error {
	action, resource := parsePermission(permission)
	if action == "" || resource == "" {
		return ErrEmptyPermission
	}
	userRoles, err := a.getRoles(ctx, userID, projectID)
	if err != nil {
		return errors.Join(ErrFailedToGetRoles, err)
	}
	if userRoles == nil {
		return ErrPermissionDenied
	}
	roleMap := NewRoles(userRoles)
	hasPermission := a.permissions.HasPermission(roleMap, action, resource)
	if !hasPermission {
		return ErrPermissionDenied
	}
	return nil
}
