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

type Secrets struct {
	AccessToken    string    `json:"access_token"`
	URN            string    `json:"urn"`
	TokenExpiresAt time.Time `json:"token_expires_at"`
}

type Linkedin struct {
	ID          string
	SecretStr   string
	userSecrets Secrets
	encrypter   encrypting.Encrypter
}

func NewLinkedin(secrets string, e encrypting.Encrypter) *Linkedin {
	l := &Linkedin{
		ID:        "linkedin",
		SecretStr: secrets,
		encrypter: e,
	}
	err := l.validateSecrets(secrets)
	if err != nil {
		fmt.Println("Error validating secrets:", err)
	}

	return l
}

func (l *Linkedin) ValidatePost(ctx context.Context, pp *post.PublishPost, media []*media.Media) error {
	posterFactory := NewLinkedinPosterFactory()
	poster, err := posterFactory.NewPoster(pp, l.userSecrets)
	if err != nil {
		return err
	}
	return poster.Validate(ctx, pp, media)
}

func (l *Linkedin) Publish(ctx context.Context, pp *post.PublishPost, media []*media.Media) error {
	fmt.Printf("Publishing post %s to LinkedIn\n", pp.ID)
	fmt.Println("Post ID:", pp.ID)
	for _, m := range media {
		fmt.Println("Media Name:", m.Filename)
	}
	posterFactory := NewLinkedinPosterFactory()
	poster, err := posterFactory.NewPoster(pp, l.userSecrets)
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

func (l *Linkedin) Authenticate(ctx context.Context, params any) (string, time.Time, error) {
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		return "", time.Time{}, errors.New("params must be a map[string]interface{}")
	}

	// Get code from params map
	codeVal, ok := paramsMap["code"]
	if !ok {
		return "", time.Time{}, errors.New("code is required in params")
	}

	// Type assert code to string
	codeStr, ok := codeVal.(string)
	if !ok {
		return "", time.Time{}, errors.New("code must be a string")
	}

	fmt.Println("Code", codeStr)
	clientID := os.Getenv("LINKEDIN_CLIENT_ID")
	clientSecret := os.Getenv("LINKEDIN_CLIENT_SECRET")
	redirectURI := os.Getenv("LINKEDIN_REDIRECT_URI")

	// Prepare form data
	formData := url.Values{}
	formData.Set("grant_type", "authorization_code")
	formData.Set("code", codeStr)
	formData.Set("client_id", clientID)
	formData.Set("client_secret", clientSecret)
	formData.Set("redirect_uri", redirectURI)

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

	userReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.linkedin.com/v2/userinfo",
		nil,
	)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	// Set auth header with the new access token
	setHeaders(userReq, tokenResp.AccessToken)

	// Send request
	userResp, err := client.Do(userReq)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to get user info: %w", err)
	}
	defer userResp.Body.Close()

	// Handle non-200 responses
	if userResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(userResp.Body)
		return "", time.Time{}, fmt.Errorf("userinfo request failed with status %d: %s", userResp.StatusCode, string(body))
	}

	// Parse user info response
	var userInfo struct {
		Sub string `json:"sub"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&userInfo); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to decode user info: %w", err)
	}

	urn := fmt.Sprintf("urn:li:person:%s", userInfo.Sub)

	// Store access token
	_, err = l.addSecret("access_token", tokenResp.AccessToken)
	if err != nil {
		return "", time.Time{}, err
	}

	// Store URN
	_, err = l.addSecret("urn", urn)
	if err != nil {
		return "", time.Time{}, err
	}

	// Store expiration
	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).UTC()
	expiresAtStr := expiresAt.Format(time.RFC3339)
	secretStr, err := l.addSecret("token_expires_at", expiresAtStr)
	if err != nil {
		return "", time.Time{}, err
	}

	return secretStr, expiresAt, nil
}

func (l *Linkedin) addSecret(key, secret string) (string, error) {
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

func (l *Linkedin) validateSecrets(secrets string) error {
	if secrets == "" || secrets == "empty" {
		return nil
	}
	var s Secrets
	err := l.encrypter.DecryptJSON(secrets, &s)
	if err != nil {
		return err
	}
	l.userSecrets = s
	return nil
}
