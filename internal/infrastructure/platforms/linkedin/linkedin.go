package linkedin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

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
	AccessToken    string    `json:"access_token"`
	URN            string    `json:"urn"`
	TokenExpiresAt time.Time `json:"token_expires_at"`
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
	case "token_expires_at":
		t, err := time.Parse(time.RFC3339, secret)
		if err != nil {
			return "", err
		}
		l.userSecrets.TokenExpiresAt = t
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
	for _, m := range media {
		fmt.Println("Media Name:", m.Filename)
	}
	posterFactory := NewPosterFactory()
	poster, err := posterFactory.NewPoster(pp, l.userSecrets, l.platformSecrets)
	if err != nil {
		return err
	}
	return poster.Post(ctx, pp, media)
}

type linkedinOAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

func (l *Linkedin) Authenticate(ctx context.Context, code string) (string, time.Time, error) {
	clientID := os.Getenv("LINKEDIN_CLIENT_ID")
	clientSecret := os.Getenv("LINKEDIN_CLIENT_SECRET")
	redirectURI := os.Getenv("LINKEDIN_REDIRECT_URI")

	// Prepare form data
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", code)
	formData.Set("client_id", clientID)
	formData.Set("client_secret", clientSecret)
	formData.Set("redirect_uri", redirectURI)

	fmt.Println(">>>>>>>>")
	fmt.Println("code:", code)
	fmt.Println("client_id:", clientID)
	fmt.Println("client_secret:", clientSecret)
	fmt.Println("redirect_uri:", redirectURI)

	// Create request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://www.linkedin.com/oauth/v2/accessToken",
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", time.Time{}, fmt.Errorf("oauth request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp linkedinOAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to decode response: %w", err)
	}

	// return encrypted secrets
	_, err = l.AddUserSecret("access_token", tokenResp.AccessToken)
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	expiresAtStr := expiresAt.Format(time.RFC3339)
	secretStr, err := l.AddUserSecret("token_expires_at", expiresAtStr)
	if err != nil {
		return "", time.Time{}, err
	}

	return secretStr, expiresAt, nil
}
