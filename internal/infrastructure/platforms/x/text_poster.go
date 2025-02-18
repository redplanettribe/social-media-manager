package x

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redplanettribe/social-media-manager/internal/domain/media"
	"github.com/redplanettribe/social-media-manager/internal/domain/post"
)

type TextPoster struct {
	secrets Secrets
}

func NewTextPoster(s Secrets) *TextPoster {
	return &TextPoster{
		secrets: s,
	}
}

func (tp *TextPoster) Validate(ctx context.Context, pp *post.PublishPost, m []*media.Media) error {
	if pp == nil {
		return errors.New("publish post is nil")
	}
	if tp.secrets.Token == "" {
		return errors.New("user access token is not set")
	}
	if tp.secrets.TokenSecret == "" {
		return errors.New("user token verifier is not set")
	}
	if pp.TextContent == "" {
		return errors.New("text content is empty")
	}
	for _, media := range m {
		if media.Size > 5242880 {
			return errors.New(fmt.Sprintf("media size exceeds 5MB: %s", media.Filename))
		}
	}

	return nil
}

func (tp *TextPoster) Post(ctx context.Context, pp *post.PublishPost, _ []*media.Media) error {
	if err := tp.Validate(ctx, pp, nil); err != nil {
		return err
	}

	// Create request body
	body := map[string]interface{}{
		"text":           pp.TextContent,
		"reply_settings": "following",
		"nullcast":       false,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal tweet body: %w", err)
	}

	// Create OAuth parameters
	oauthParams := map[string]string{
		"oauth_consumer_key":     os.Getenv("X_API_KEY"),
		"oauth_nonce":            uuid.New().String()[:32],
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        fmt.Sprintf("%d", time.Now().Unix()),
		"oauth_token":            tp.secrets.Token,
		"oauth_version":          "1.0",
	}

	// Create signature base string
	baseURL := "https://api.x.com/2/tweets"
	method := "POST"
	baseStr := createSignatureBaseString(method, baseURL, oauthParams)

	// Create signing key
	signingKey := fmt.Sprintf("%s&%s",
		url.QueryEscape(os.Getenv("X_API_SECRET")),
		url.QueryEscape(tp.secrets.TokenSecret),
	)

	// Generate signature
	h := hmac.New(sha1.New, []byte(signingKey))
	h.Write([]byte(baseStr))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	oauthParams["oauth_signature"] = url.QueryEscape(signature)

	// Create Authorization header
	var headerParts []string
	for key, value := range oauthParams {
		headerParts = append(headerParts, fmt.Sprintf("%s=\"%s\"", key, value))
	}
	authHeader := "OAuth " + strings.Join(headerParts, ", ")

	// Create request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		baseURL,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to create tweet request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send tweet request: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-201 responses
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tweet failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tweetResponse struct {
		Data struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tweetResponse); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Tweet published successfully with ID: %s\n", tweetResponse.Data.ID)
	return nil
}
