package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

type TextPoster struct {
	secrets Secrets
}

func NewTextPoster(s Secrets) *TextPoster {
	return &TextPoster{
		secrets: s,
	}
}

func (tp *TextPoster) Post(ctx context.Context, pp *post.PublishPost, _ []*media.Media) error {
	// Validate input
	if pp == nil {
		return errors.New("publish post is nil")
	}
	if tp.secrets.AccessToken == "" {
		return errors.New("user access token is not set")
	}
	if tp.secrets.URN == "" {
		return errors.New("user URN is not set")
	}

	body := map[string]interface{}{
		"author":     tp.secrets.URN,
		"commentary": pp.TextContent,
		"visibility": "PUBLIC",
		"distribution": map[string]interface{}{
			"feedDistribution":               "MAIN_FEED",
			"targetEntities":                 []string{},
			"thirdPartyDistributionChannels": []string{},
		},
		"lifecycleState":            "PUBLISHED",
		"isReshareDisabledByAuthor": false,
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal LinkedIn post body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linkedin.com/rest/posts", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create LinkedIn post request: %w", err)
	}

	setHeaders(req, tp.secrets.AccessToken)

	// Initialize an HTTP client with a timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Send the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send LinkedIn post request: %w", err)
	}
	defer resp.Body.Close()

	// Handle the response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var respBody map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&respBody); decodeErr != nil {
			return fmt.Errorf("LinkedIn API responded with status %d", resp.StatusCode)
		}
		return fmt.Errorf("LinkedIn API responded with status %d: %v", resp.StatusCode, respBody)
	}

	fmt.Println("Text post published successfully to LinkedIn.")
	return nil
}
