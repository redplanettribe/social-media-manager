package x

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redplanettribe/social-media-manager/internal/domain/media"
	"github.com/redplanettribe/social-media-manager/internal/domain/post"
)

// processingInfo represents asynchronous processing data returned during FINALIZE and STATUS.
type processingInfo struct {
	State           string `json:"state"`
	CheckAfterSecs  int    `json:"check_after_secs"`
	ProgressPercent int    `json:"progress_percent,omitempty"`
}

// MediaPoster implements an upload client for media to the X API.
type MediaPoster struct {
	secrets Secrets
}

// NewMediaPoster returns a new MediaPoster.
func NewMediaPoster(s Secrets) *MediaPoster {
	return &MediaPoster{
		secrets: s,
	}
}

// Post orchestrates media upload (INIT → APPEND → FINALIZE → STATUS) then creates a tweet with the returned media IDs.
func (ip *MediaPoster) Post(ctx context.Context, pp *post.PublishPost, mediaList []*media.Media) error {
	if err := ip.Validate(ctx, pp, mediaList); err != nil {
		return err
	}
	var mediaIDs []string
	for _, m := range mediaList {
		id, err := ip.uploadMedia(ctx, m)
		if err != nil {
			return err
		}
		mediaIDs = append(mediaIDs, id)
	}
	return ip.createTweet(ctx, pp, mediaIDs)
}

// Validate verifies required fields.
func (ip *MediaPoster) Validate(ctx context.Context, pp *post.PublishPost, mediaList []*media.Media) error {
	if pp == nil {
		return fmt.Errorf("publish post is nil")
	}
	if ip.secrets.Token == "" {
		return fmt.Errorf("user access token not set")
	}
	if len(mediaList) == 0 {
		return fmt.Errorf("no media to upload")
	}
	return nil
}

// uploadMedia performs the complete media upload workflow.
func (ip *MediaPoster) uploadMedia(ctx context.Context, m *media.Media) (string, error) {
	// Step 1: INIT upload session.
	mediaID, err := ip.initUpload(ctx, m)
	if err != nil {
		return "", err
	}

	// Step 2: APPEND media chunks.
	if err := ip.appendUpload(ctx, mediaID, m, 0); err != nil {
		return "", err
	}

	// Step 3: FINALIZE upload.
	finalMediaID, procInfo, err := ip.finalizeUpload(ctx, mediaID)
	if err != nil {
		return "", err
	}

	// Step 4: Poll STATUS if processing_info is returned.
	if procInfo != nil {
		if err := ip.pollUploadStatus(ctx, finalMediaID); err != nil {
			return "", err
		}
	}
	return finalMediaID, nil
}

// initUpload uses the INIT command to create an upload session.
func (ip *MediaPoster) initUpload(ctx context.Context, m *media.Media) (string, error) {
	fmt.Println("Initiating upload for", m.Filename)
	urlStr := "https://api.x.com/2/media/upload"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Set INIT required fields.
	_ = writer.WriteField("command", "INIT")
	_ = writer.WriteField("total_bytes", fmt.Sprintf("%d", len(m.Data)))
	// Set media type based on the file's format.
	mediaType := "image/" + m.Format
	if m.IsVideo() {
		mediaType = "video/" + m.Format
	}
	_ = writer.WriteField("media_type", mediaType)
	// For videos, use amplify_video; for images, use tweet_image.
	category := "tweet_image"
	if m.IsVideo() {
		category = "amplify_video"
	}
	_ = writer.WriteField("media_category", category)
	writer.Close()

	authHeader := ip.buildOAuthHeader(http.MethodPost, urlStr, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, body)
	if err != nil {
		return "", fmt.Errorf("failed to create INIT request: %w", err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send INIT request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("INIT failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var initResp struct {
		Data struct {
			ID               string `json:"id"`
			MediaKey         string `json:"media_key"`
			ExpiresAfterSecs int    `json:"expires_after_secs"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		return "", fmt.Errorf("failed to decode INIT response: %w", err)
	}
	return initResp.Data.ID, nil
}

// appendUpload sends one or more APPEND requests to upload all file chunks.
func (ip *MediaPoster) appendUpload(ctx context.Context, mediaID string, m *media.Media, segmentIndex int) error {
	fmt.Println("Appending media for", m.Filename)
	urlStr := "https://api.x.com/2/media/upload"

	// If video, split file into 4 MB chunks.
	if m.IsVideo() {
		const chunkSize = 3 * 1024 * 1024 // 3 MB
		totalSize := len(m.Data)
		segmentIndex = 0
		for offset := 0; offset < totalSize; offset += chunkSize {
			end := offset + chunkSize
			if end > totalSize {
				end = totalSize
			}
			chunk := m.Data[offset:end]
			if err := ip.sendAppendRequest(ctx, urlStr, mediaID, m.Filename, segmentIndex, chunk); err != nil {
				return err
			}
			fmt.Printf("Appending segment %d for %s\n", segmentIndex, m.Filename)
			segmentIndex++
		}
	} else {
		// For non-videos, one APPEND is sufficient.
		if err := ip.sendAppendRequest(ctx, urlStr, mediaID, m.Filename, 0, m.Data); err != nil {
			return err
		}
	}
	return nil
}

// sendAppendRequest sends an individual APPEND request for a file chunk.
func (ip *MediaPoster) sendAppendRequest(ctx context.Context, urlStr, mediaID, filename string, segmentIndex int, data []byte) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("command", "APPEND")
	_ = writer.WriteField("media_id", mediaID)
	_ = writer.WriteField("segment_index", strconv.Itoa(segmentIndex))
	part, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return fmt.Errorf("failed to create media file field: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("failed to write media data: %w", err)
	}
	writer.Close()

	authHeader := ip.buildOAuthHeader(http.MethodPost, urlStr, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, body)
	if err != nil {
		return fmt.Errorf("failed to create APPEND request: %w", err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send APPEND request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("APPEND failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}

// finalizeUpload sends the FINALIZE command and returns the media ID.
func (ip *MediaPoster) finalizeUpload(ctx context.Context, mediaID string) (string, *processingInfo, error) {
	fmt.Println("Finalizing upload for", mediaID)
	urlStr := "https://api.x.com/2/media/upload"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("command", "FINALIZE")
	_ = writer.WriteField("media_id", mediaID)
	writer.Close()

	authHeader := ip.buildOAuthHeader(http.MethodPost, urlStr, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create FINALIZE request: %w", err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to send FINALIZE request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("FINALIZE failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var finResp struct {
		Data struct {
			ID               string          `json:"id"`
			MediaKey         string          `json:"media_key"`
			Size             int             `json:"size"`
			ExpiresAfterSecs int             `json:"expires_after_secs"`
			ProcessingInfo   *processingInfo `json:"processing_info,omitempty"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&finResp); err != nil {
		return "", nil, fmt.Errorf("failed to decode FINALIZE response: %w", err)
	}

	fmt.Println("Finalize response:", finResp)

	// Use the "id" field (a 64-bit numeric string) directly for tweet creation.
	finalMediaID := finResp.Data.ID

	return finalMediaID, finResp.Data.ProcessingInfo, nil
}

// pollUploadStatus polls until the video processing is complete.
func (ip *MediaPoster) pollUploadStatus(ctx context.Context, mediaID string) error {
	fmt.Println("Polling upload status for", mediaID)
	baseURL := "https://api.x.com/2/media/upload"
	params := url.Values{}
	params.Add("command", "STATUS")
	params.Add("media_id", mediaID)
	fullURL := baseURL + "?" + params.Encode()

	// Parse the URL to extract query parameters.
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}
	extraParams := make(map[string]string)
	for key, values := range parsedURL.Query() {
		if len(values) > 0 {
			extraParams[key] = values[0]
		}
	}
	// Build the base URL without the query before signing.
	signingURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path)

	client := &http.Client{Timeout: 10 * time.Second}
	for {
		authHeader := ip.buildOAuthHeader(http.MethodGet, signingURL, extraParams)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create STATUS request: %w", err)
		}
		req.Header.Set("Authorization", authHeader)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to send STATUS request: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		fmt.Println("Raw response:", string(respBody))

		var statusResp struct {
			Data struct {
				ID             string          `json:"id"`
				MediaKey       string          `json:"media_key"`
				ProcessingInfo *processingInfo `json:"processing_info"`
			} `json:"data"`
		}

		if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&statusResp); err != nil {
			return fmt.Errorf("failed to decode STATUS response: %w", err)
		}
		fmt.Printf("Decoded response: %+v\n", statusResp)

		if statusResp.Data.ProcessingInfo != nil {
			state := statusResp.Data.ProcessingInfo.State
			if state == "succeeded" {
				return nil
			} else if state == "failed" {
				return fmt.Errorf("media processing failed")
			}
			waitTime := time.Duration(statusResp.Data.ProcessingInfo.CheckAfterSecs) * time.Second
			if waitTime == 0 {
				waitTime = 5 * time.Second
			}
			time.Sleep(waitTime)
		} else {
			// Assume the media is ready if no processing info is returned.
			return nil
		}
	}
}

// createTweet sends a tweet including the provided media IDs.
func (ip *MediaPoster) createTweet(ctx context.Context, pp *post.PublishPost, mediaIDs []string) error {
	fmt.Println("Creating tweet")
	fmt.Println("Media IDs:", mediaIDs)
	tweetBody := map[string]interface{}{
		"text": pp.TextContent,
		"media": map[string]interface{}{
			"media_ids": mediaIDs,
		},
		"reply_settings": "following",
	}
	tweetJSON, err := json.Marshal(tweetBody)
	if err != nil {
		return fmt.Errorf("failed to marshal tweet body: %w", err)
	}

	baseURL := "https://api.x.com/2/tweets"
	authHeader := ip.buildOAuthHeader(http.MethodPost, baseURL, nil)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewBuffer(tweetJSON))
	if err != nil {
		return fmt.Errorf("failed to create tweet request: %w", err)
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send tweet request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tweet failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// buildOAuthHeader builds an OAuth 1.0a header for signing requests.
func (ip *MediaPoster) buildOAuthHeader(method, baseURL string, extraParams map[string]string) string {
	// Basic OAuth parameters.
	oauthParams := map[string]string{
		"oauth_consumer_key":     url.QueryEscape(os.Getenv("X_API_KEY")),
		"oauth_nonce":            url.QueryEscape(uuid.New().String()[:32]),
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        fmt.Sprintf("%d", time.Now().Unix()),
		"oauth_token":            url.QueryEscape(ip.secrets.Token),
		"oauth_version":          "1.0",
	}
	// Merge any extra parameters.
	for k, v := range extraParams {
		oauthParams[k] = v
	}

	// Create the signature base string.
	baseStr := createSignatureBaseString(method, baseURL, oauthParams)
	// Signing key combines API secret with token secret.
	signingKey := fmt.Sprintf("%s&%s", url.QueryEscape(os.Getenv("X_API_SECRET")), url.QueryEscape(ip.secrets.TokenSecret))
	h := hmac.New(sha1.New, []byte(signingKey))
	h.Write([]byte(baseStr))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	oauthParams["oauth_signature"] = url.QueryEscape(signature)

	// Build the Authorization header.
	var headerParts []string
	for key, value := range oauthParams {
		headerParts = append(headerParts, fmt.Sprintf("%s=\"%s\"", key, value))
	}
	return "OAuth " + strings.Join(headerParts, ", ")
}
