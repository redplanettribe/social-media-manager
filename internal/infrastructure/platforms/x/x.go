package x

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redplanettribe/social-media-manager/internal/domain/media"
	"github.com/redplanettribe/social-media-manager/internal/domain/post"
	"github.com/redplanettribe/social-media-manager/internal/infrastructure/encrypting"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

type Secrets struct {
	Token       string `json:"access_token"`
	TokenSecret string `json:"token_verifier"`
}

type X struct {
	ID          string
	SecretStr   string
	userSecrets Secrets
	encrypter   encrypting.Encrypter
}

func NewX(secrets string, e encrypting.Encrypter) *X {
	x := &X{
		ID:        "x",
		SecretStr: secrets,
		encrypter: e,
	}
	err := x.validateSecrets(secrets)
	if err != nil {
		fmt.Println("Error validating secrets:", err)
	}

	return x
}

func (x *X) MemberLookup(ctx context.Context, username string) (string, error) {
	return "", ErrNotImplemented
}

func (x *X) validateSecrets(secrets string) error {
	if secrets == "" || secrets == "empty" {
		return nil
	}
	var s Secrets
	err := x.encrypter.DecryptJSON(secrets, &s)
	if err != nil {
		return err
	}
	x.userSecrets = s
	return nil
}

func (x *X) Authenticate(ctx context.Context, params any) (string, time.Time, error) {
	paramsMap, ok := params.(map[string]interface{})
	fmt.Println("Params:", paramsMap)
	if !ok {
		return "", time.Time{}, errors.New("params must be a map[string]interface{}")
	}

	// Get oauth_token and verifier from params
	oauthToken, ok := paramsMap["oauth_token"].(string)
	if !ok {
		return "", time.Time{}, errors.New("oauth_token must be a string")
	}

	oauthVerifier, ok := paramsMap["oauth_verifier"].(string)
	if !ok {
		return "", time.Time{}, errors.New("oauth_verifier must be a string")
	}

	// Create request for access token
	formData := url.Values{}
	formData.Set("oauth_token", oauthToken)
	formData.Set("oauth_verifier", oauthVerifier)

	// Create OAuth parameters
	reqParams := map[string]string{
		"oauth_consumer_key":     os.Getenv("X_API_KEY"),
		"oauth_nonce":            uuid.New().String()[:32],
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        fmt.Sprintf("%d", time.Now().Unix()),
		"oauth_token":            oauthToken,
		"oauth_verifier":         oauthVerifier,
		"oauth_version":          "1.0",
	}

	// Create signature base string
	baseStr := createSignatureBaseString("POST", "https://api.x.com/oauth/access_token", reqParams)

	// Create signing key
	signingKey := fmt.Sprintf("%s&%s",
		url.QueryEscape(os.Getenv("X_API_SECRET")),
		url.QueryEscape(oauthVerifier),
	)

	// Generate signature
	h := hmac.New(sha1.New, []byte(signingKey))
	h.Write([]byte(baseStr))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	reqParams["oauth_signature"] = url.QueryEscape(signature)

	// Create Authorization header
	var headerParts []string
	for key, value := range reqParams {
		headerParts = append(headerParts, fmt.Sprintf("%s=\"%s\"", key, value))
	}
	authHeader := "OAuth " + strings.Join(headerParts, ", ")

	// Create request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.x.com/oauth/access_token",
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", authHeader)
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to read response: %w", err)
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Save secrets
	x.userSecrets = Secrets{
		Token:       values.Get("oauth_token"),
		TokenSecret: values.Get("oauth_token_secret"),
	}

	// Encrypt secrets
	secretStr, err := x.encrypter.EncryptJSON(x.userSecrets)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to encrypt secrets: %w", err)
	}

	// X tokens don't expire, so we return a far future time
	expiresAt := time.Now().AddDate(100, 0, 0)

	return secretStr, expiresAt, nil
}

func createSignatureBaseString(method, baseURL string, params map[string]string) string {
	// Sort parameters alphabetically
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build parameter string
	var paramPairs []string
	for _, k := range keys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", k, params[k]))
	}
	paramString := strings.Join(paramPairs, "&")

	// Create signature base string
	return fmt.Sprintf("%s&%s&%s",
		url.QueryEscape(method),
		url.QueryEscape(baseURL),
		url.QueryEscape(paramString),
	)
}

func (x *X) ValidatePost(ctx context.Context, pp *post.PublishPost, media []*media.Media) error {
	posterFactory := NewXPosterFactory()
	poster, err := posterFactory.NewPoster(pp, x.userSecrets)
	if err != nil {
		return err
	}
	return poster.Validate(ctx, pp, media)
}

func (x *X) Publish(ctx context.Context, pp *post.PublishPost, media []*media.Media) error {
	for _, m := range media {
		fmt.Println("Media Name:", m.Filename)
	}
	posterFactory := NewXPosterFactory()
	poster, err := posterFactory.NewPoster(pp, x.userSecrets)
	if err != nil {
		return err
	}
	return poster.Post(ctx, pp, media)
}
