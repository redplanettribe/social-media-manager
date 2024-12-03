package auth

import (
	"context"

	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/session"
)

type Authenticator interface {
	Authenticate(ctx context.Context, sessionID string) error
}

type SessionAuthenticator struct {
	sessionManager session.Manager
}

func NewAuthenticator(sessionManager session.Manager) Authenticator {
	return &SessionAuthenticator{
		sessionManager: sessionManager,
	}
}

func (a *SessionAuthenticator) Authenticate(ctx context.Context, sessionID string) error {
	_, err := a.sessionManager.ValidateSession(ctx, sessionID)
	if err != nil {
		return err
	}
	return nil
}
