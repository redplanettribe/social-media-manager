package platforms

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
type Secrets struct {
	AccessToken string `json:"access_token"`
	UserUrn     string `json:"user_urn"`
}

type Linkedin struct {
	ID        string
	SecretStr string
	secrets   Secrets
	encrypter encrypting.Encrypter
}



func NewLinkedin(secrets string, e encrypting.Encrypter) *Linkedin {
	return &Linkedin{
		ID:        "linkedin",
		SecretStr: secrets,
		encrypter: e,
	}
}


func (l *Linkedin) Publish(ctx context.Context, post *post.PublishPost, media []*media.Media) error {
	// Publish to Linkedin
	fmt.Println("Publishing to Linkedin: ", post.Title, post.ID)
	time.Sleep(1 * time.Second)
	fmt.Println("Published to Linkedin")
	return nil
}

func (l *Linkedin) ValidateSecrets(secrets string) error {
	var s Secrets
	err := l.encrypter.DecryptJSON(secrets, &s)
	if err != nil {
		return err
	}

	l.secrets = s

	return nil
}

func (l *Linkedin) AddSecret(key, secret string) (string, error) {
	switch key {
	case "access_token":
		l.secrets.AccessToken = secret
	case "user_urn":
		l.secrets.UserUrn = secret
	default:
		return "", errors.New("invalid key")
	}

	newSecretStr, err := l.encrypter.EncryptJSON(l.secrets)
	if err != nil {
		return "", err
	}

	return newSecretStr, nil
}

