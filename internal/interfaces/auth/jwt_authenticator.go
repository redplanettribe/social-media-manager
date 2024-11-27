package auth

import (
	"github.com/google/uuid"
)

type Authenticator interface {
	Authenticate(tokenString string) (uuid.UUID, error)
	GenerateToken(userID uuid.UUID) (string, error)
}

type JWTAuthenticator struct {
	secretKey string
}

func NewJWTAuthenticator(secretKey string) *JWTAuthenticator {
	return &JWTAuthenticator{secretKey: secretKey}
}

func (a *JWTAuthenticator) Authenticate(tokenString string) (uuid.UUID, error) {

	return uuid.UUID{}, nil
}

func (a *JWTAuthenticator) GenerateToken(userID uuid.UUID) (string, error) {
	return "", nil
}
