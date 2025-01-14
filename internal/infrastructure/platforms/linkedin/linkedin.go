package linkedin

import (
	"context"
	"errors"
	"fmt"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/encrypting"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

type PlatformSecrets struct {
	PlaceHolderKey string `json:"placeholder_key"`
}

type UserSecrets struct {
	AccessToken string `json:"access_token"`
	URN         string `json:"urn"`
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
	case "urn":
		l.userSecrets.URN = secret
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
	if secrets == "" || secrets == "empty" {
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
	if secrets == "" || secrets == "empty" {
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

func (l *Linkedin) Publish(ctx context.Context, pp *post.PublishPost, media []*media.Media) error {
	fmt.Println("Publishing to Linkedin")
	fmt.Println("Post ID:", pp.ID)
	posterFactory := NewPosterFactory()
	poster, err := posterFactory.NewPoster(pp, l.userSecrets , l.platformSecrets)
	if err != nil {
		return err
	}
	return poster.Post(ctx, pp, media)
}
