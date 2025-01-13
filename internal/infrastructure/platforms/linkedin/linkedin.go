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
	if pp.Type == post.PostTypeUndefined{
		t, err := l.AttemptToClassifyPostType(pp, media)
		if err != nil {
			return err
		}
		pp.Type = post.PostType(t)
	}
	switch pp.Type {
	case post.PostTypeText:
		return l.publishTextPost(ctx, pp)
	case post.PostTypeMedia:
		return l.publishMediaPost(ctx, pp, media)
	case post.PostTypePoll:
		return l.publishPollPost(ctx, pp)
	case post.PostTypeUndefined:
		return fmt.Errorf("post type is undefined")
	default:
		return fmt.Errorf("unsupported post type: %s", pp.Type)
	}
}

func (l *Linkedin) AttemptToClassifyPostType(pp *post.PublishPost, media []*media.Media) (post.PostType, error) {
	//TODO: Implement this, for now it always returns PostTypeUndefined
	return post.PostTypeUndefined, nil
}

func (l *Linkedin) publishTextPost(ctx context.Context, pp *post.PublishPost) error {
	//TODO: Implement this
	return ErrNotImplemented
}

func (l *Linkedin) publishMediaPost(ctx context.Context, pp *post.PublishPost, media []*media.Media) error {
	//TODO: Implement this
	return ErrNotImplemented
}

func (l *Linkedin) publishPollPost(ctx context.Context, pp *post.PublishPost) error {
	//TODO: Implement this
	return ErrNotImplemented
}
