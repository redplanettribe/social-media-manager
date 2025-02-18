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

type DocumentPoster struct {
	secrets    Secrets
	authorURN  string
	httpClient *http.Client
}

func NewDocumentPoster(s Secrets) *DocumentPoster {
	return &DocumentPoster{
		secrets:    s,
		authorURN:  s.URN,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (dp *DocumentPoster) Validate(ctx context.Context, pp *post.PublishPost, mediaList []*media.Media) error {
	if pp == nil {
		return errors.New("publish post is nil")
	}
	if dp.secrets.AccessToken == "" {
		return errors.New("user access token is not set")
	}
	if dp.authorURN == "" {
		return errors.New("user URN is not set")
	}
	if len(mediaList) == 0 {
		return errors.New("no document to upload")
	}
	if len(mediaList) > 1 {
		return errors.New("single document upload only")
	}
	d := mediaList[0]
	if pp.Type == post.PostTypeCarousel && d.Format != "pdf" {
		return errors.New("carousel post requires a PDF document")
	}
	return nil
}

// initDocUploadReq and initDocUploadResp mirror LinkedIn’s document upload initialization process.
type initDocUploadReq struct {
	InitializeUploadRequest struct {
		Owner string `json:"owner"`
	} `json:"initializeUploadRequest"`
}

type initDocUploadResp struct {
	Value struct {
		UploadUrlExpiresAt int64  `json:"uploadUrlExpiresAt"`
		UploadUrl          string `json:"uploadUrl"`
		Document           string `json:"document"`
	} `json:"value"`
}

// Post uploads a document and creates a LinkedIn post referring to it.
func (dp *DocumentPoster) Post(ctx context.Context, pp *post.PublishPost, mediaList []*media.Media) error {
	if err := dp.Validate(ctx, pp, mediaList); err != nil {
		return err
	}
	// Step 1: Initialize document upload
	initReqBody := initDocUploadReq{}
	initReqBody.InitializeUploadRequest.Owner = dp.authorURN

	bodyData, err := json.Marshal(initReqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal init request: %w", err)
	}

	initReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linkedin.com/rest/documents?action=initializeUpload", bytes.NewBuffer(bodyData))
	if err != nil {
		return fmt.Errorf("failed to create init request: %w", err)
	}
	setHeaders(initReq, dp.secrets.AccessToken)

	initResp, err := dp.httpClient.Do(initReq)
	if err != nil {
		return fmt.Errorf("failed to send init request: %w", err)
	}
	defer initResp.Body.Close()

	if initResp.StatusCode != http.StatusOK && initResp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(initResp.Body)
		return fmt.Errorf("initialize document upload failed (%d): %s", initResp.StatusCode, string(respBody))
	}

	var initRes initDocUploadResp
	if err := json.NewDecoder(initResp.Body).Decode(&initRes); err != nil {
		return fmt.Errorf("failed to decode init response: %w", err)
	}
	if initRes.Value.UploadUrl == "" || initRes.Value.Document == "" {
		return errors.New("invalid init response: missing uploadUrl or document URN")
	}

	// Step 2: Upload the document as binary
	doc := mediaList[0]
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPut, initRes.Value.UploadUrl, bytes.NewBuffer(doc.Data))
	if err != nil {
		return fmt.Errorf("failed to create document upload request: %w", err)
	}
	setBinaryHeaders(uploadReq, dp.secrets.AccessToken)

	uploadResp, err := dp.httpClient.Do(uploadReq)
	if err != nil {
		return fmt.Errorf("failed to upload document data: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusOK && uploadResp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(uploadResp.Body)
		return fmt.Errorf("document upload failed (%d): %s", uploadResp.StatusCode, string(respBody))
	}

	// Step 3: Create the post referencing the document
	finalBody := map[string]interface{}{
		"author":     dp.authorURN,
		"commentary": pp.TextContent,
		"visibility": "PUBLIC",
		"distribution": map[string]interface{}{
			"feedDistribution":               "MAIN_FEED",
			"targetEntities":                 []string{},
			"thirdPartyDistributionChannels": []string{},
		},
		"content": map[string]interface{}{
			"media": map[string]interface{}{
				"id":    initRes.Value.Document,
				"title": doc.Filename, // or a hardcoded title
			},
		},
		"lifecycleState":            "PUBLISHED",
		"isReshareDisabledByAuthor": false,
	}

	postData, err := json.Marshal(finalBody)
	if err != nil {
		return fmt.Errorf("failed to marshal final post: %w", err)
	}

	postReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linkedin.com/rest/posts", bytes.NewBuffer(postData))
	if err != nil {
		return fmt.Errorf("failed to create final post request: %w", err)
	}
	setHeaders(postReq, dp.secrets.AccessToken)

	postResp, err := dp.httpClient.Do(postReq)
	if err != nil {
		return fmt.Errorf("failed to send final post request: %w", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusOK && postResp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(postResp.Body)
		return fmt.Errorf("create document post failed (%d): %s", postResp.StatusCode, string(respBody))
	}

	fmt.Println("Document post published successfully to LinkedIn.")
	return nil
}
