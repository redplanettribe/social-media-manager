package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/redplanettribe/social-media-manager/internal/domain/media"
	"github.com/redplanettribe/social-media-manager/internal/domain/post"
)

// ImagePoster only deals with image uploads.
type ImagePoster struct {
	secrets    Secrets
	httpClient *http.Client
	authorURN  string
}

func NewImagePoster(s Secrets) *ImagePoster {
	return &ImagePoster{
		secrets:    s,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		authorURN:  s.URN,
	}
}

func (ip *ImagePoster) Validate(ctx context.Context, pp *post.PublishPost, mediaList []*media.Media) error {
	if pp == nil {
		return errors.New("publish post is nil")
	}
	if ip.secrets.AccessToken == "" {
		return errors.New("user access token is not set")
	}
	if ip.authorURN == "" {
		return errors.New("user URN is not set")
	}
	if len(mediaList) == 0 {
		return errors.New("no images to upload")
	}
	if len(mediaList) > 1 {
		return errors.New("single image upload only")
	}
	return nil
}

// initUploadRequest contains the request body for initializing an image upload.
type initUploadRequest struct {
	InitializeUploadRequest struct {
		Owner string `json:"owner"`
	} `json:"initializeUploadRequest"`
}

// initUploadResponse contains the response for initializing an image upload.
type initUploadResponse struct {
	Value struct {
		UploadUrlExpiresAt int64  `json:"uploadUrlExpiresAt"`
		UploadUrl          string `json:"uploadUrl"`
		Image              string `json:"image"`
	} `json:"value"`
}

func (ip *ImagePoster) Post(ctx context.Context, pp *post.PublishPost, mediaList []*media.Media) error {
	if err := ip.Validate(ctx, pp, mediaList); err != nil {
		return err
	}

	m := mediaList[0]

	// 1. Initialize Upload
	initReqBody := initUploadRequest{}
	initReqBody.InitializeUploadRequest.Owner = ip.authorURN

	bodyBytes, err := json.Marshal(initReqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal initialize upload request: %w", err)
	}

	initReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linkedin.com/rest/images?action=initializeUpload", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create initialize upload request: %w", err)
	}
	setHeaders(initReq, ip.secrets.AccessToken)

	resp, err := ip.httpClient.Do(initReq)
	if err != nil {
		return fmt.Errorf("failed to send initialize upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("initialize upload failed with status %d", resp.StatusCode)
	}

	var initResp initUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
		return fmt.Errorf("failed to decode initialize upload response: %w", err)
	}

	uploadURL := initResp.Value.UploadUrl
	imageURN := initResp.Value.Image

	if uploadURL == "" || imageURN == "" {
		return errors.New("invalid initialize upload response: missing uploadURL or image URN")
	}

	// 2. Upload the image binary to the returned uploadUrl
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURL, bytes.NewBuffer(m.Data))
	if err != nil {
		return fmt.Errorf("failed to create image upload request: %w", err)
	}
	setBinaryHeaders(uploadReq, ip.secrets.AccessToken)

	uploadResp, err := ip.httpClient.Do(uploadReq)
	if err != nil {
		return fmt.Errorf("failed to send image upload request: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK && uploadResp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(uploadResp.Body)
		return fmt.Errorf("image upload failed with status %d: %s", uploadResp.StatusCode, string(respBody))
	}

	// 3. Create the post referencing the uploaded image
	finalBody := map[string]interface{}{
		"author":     ip.authorURN,
		"commentary": pp.TextContent,
		"visibility": "PUBLIC",
		"distribution": map[string]interface{}{
			"feedDistribution":               "MAIN_FEED",
			"targetEntities":                 []string{},
			"thirdPartyDistributionChannels": []string{},
		},
		"content": map[string]interface{}{
			"media": map[string]interface{}{
				"altText": m.AltText,
				"id":      imageURN,
			},
		},
		"lifecycleState":            "PUBLISHED",
		"isReshareDisabledByAuthor": false,
	}

	postPayload, err := json.Marshal(finalBody)
	if err != nil {
		return fmt.Errorf("failed to marshal final post payload: %w", err)
	}

	postReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linkedin.com/rest/posts", bytes.NewBuffer(postPayload))
	if err != nil {
		return fmt.Errorf("failed to create final post request: %w", err)
	}
	setHeaders(postReq, ip.secrets.AccessToken)

	postResp, err := ip.httpClient.Do(postReq)
	if err != nil {
		return fmt.Errorf("failed to send final post request: %w", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusOK && postResp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(postResp.Body)
		return fmt.Errorf("create post failed with status %d: %s", postResp.StatusCode, string(respBody))
	}

	fmt.Println("Image post published successfully to LinkedIn.")
	return nil
}
