package linkedin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/encrypting"
)

// placeholder fields for secrets
type PlatformSecrets struct {
	PlaceHolderKey string `json:"placeholder_key"`
}

type UserSecrets struct {
	AccessToken string `json:"access_token"`
}

type Linkedin struct {
	ID              string
	SecretStr       string
	platformSecrets PlatformSecrets
	userSecrets     UserSecrets
	encrypter       encrypting.Encrypter
}

func NewLinkedin(secrets string, e encrypting.Encrypter) *Linkedin {
	return &Linkedin{
		ID:        "linkedin",
		SecretStr: secrets,
		encrypter: e,
	}
}

func (l *Linkedin) Publish(ctx context.Context, post *post.PublishPost, media []*media.Media) error {
	fmt.Println("access_token", l.userSecrets.AccessToken)
	fmt.Println("platform secrets", l.platformSecrets)


	fmt.Println("Publishing to Linkedin: ", post.Title, post.ID)
	time.Sleep(1 * time.Second)
	fmt.Println("Published to Linkedin")
	return nil
}



func (l *Linkedin) AddPlatformSecret(key, secret string) (string, error) {
	switch key {
	case "placeholder_key":
		l.platformSecrets.PlaceHolderKey = secret
	default:
		return "", errors.New("invalid key")
	}

	newSecretStr, err := l.encrypter.EncryptJSON(l.platformSecrets)
	if err != nil {
		return "", err
	}

	return newSecretStr, nil
}



func (l *Linkedin) AddUserSecret(key, secret string) (string, error) {
	switch key {
	case "access_token":
		l.userSecrets.AccessToken = secret
	default:
		return "", errors.New("invalid key")
	}

	newSecretStr, err := l.encrypter.EncryptJSON(l.userSecrets)
	if err != nil {
		return "", err
	}

	return newSecretStr, nil
}

func (l *Linkedin) ValidateUserSecrets(secrets string) error {
	if secrets == "" || secrets == "empty"{
		return nil
	}
	var s UserSecrets
	err := l.encrypter.DecryptJSON(secrets, &s)
	if err != nil {
		return err
	}
	l.userSecrets = s
	return nil
}

func (l *Linkedin) ValidatePlatformSecrets(secrets string) error {
	if secrets == "" || secrets == "empty"{
		return nil
	}
	var s PlatformSecrets
	err := l.encrypter.DecryptJSON(secrets, &s)
	if err != nil {
		return err
	}
	l.platformSecrets = s
	return nil
}
