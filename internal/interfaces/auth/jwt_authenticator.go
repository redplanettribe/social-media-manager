package auth

import (
	"log"

	"github.com/google/uuid"
)

var testUserID = uuid.New()

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
	log.Printf("Authenticating token: %s", tokenString)
	return testUserID, nil
}

func (a *JWTAuthenticator) GenerateToken(userID uuid.UUID) (string, error) {
	return "", nil
}
