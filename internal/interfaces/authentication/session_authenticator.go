package authentication

import (
	"context"

	"github.com/redplanettribe/social-media-manager/internal/infrastructure/session"
)

type Authenticator interface {
	Authenticate(ctx context.Context, sessionID, fingerprint string) (*session.Session, error)
}

type SessionAuthenticator struct {
	sessionManager session.Manager
}

func NewAuthenticator(sessionManager session.Manager) Authenticator {
	return &SessionAuthenticator{
		sessionManager: sessionManager,
	}
}

func (a *SessionAuthenticator) Authenticate(ctx context.Context, sessionID, fingerprint string) (*session.Session, error) {
	session, err := a.sessionManager.ValidateSession(ctx, sessionID, fingerprint)
	if err != nil {
		return nil, err
	}
	return session, nil
}
