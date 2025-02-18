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

type VideoPoster struct {
	secrets    Secrets
	httpClient *http.Client
	authorURN  string
}

func NewVideoPoster(s Secrets) *VideoPoster {
	return &VideoPoster{
		secrets:    s,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		authorURN:  s.URN,
	}
}

func (vp *VideoPoster) Validate(ctx context.Context, pp *post.PublishPost, mediaList []*media.Media) error {
	if pp == nil {
		return errors.New("publish post is nil")
	}
	if vp.secrets.AccessToken == "" {
		return errors.New("user access token is not set")
	}
	if vp.authorURN == "" {
		return errors.New("user URN is not set")
	}
	if len(mediaList) != 1 {
		return errors.New("exactly one video is required")
	}

	m := mediaList[0]
	if !m.IsVideo() {
		return errors.New("provided media is not a video")
	}
	if m.Size > 500*1024*1024 {
		return errors.New("video file size is too large")
	}
	if m.Length > 30*60 {
		return errors.New("video length is too long")
	}
	if m.Length < 3 {
		return errors.New("video length is too short")
	}
	if m.Format != "mp4" {
		return errors.New("video format is not mp4")
	}

	return nil
}

type initVideoUploadReq struct {
	InitializeUploadRequest struct {
		Owner           string `json:"owner"`
		FileSizeBytes   int64  `json:"fileSizeBytes"`
		UploadCaptions  bool   `json:"uploadCaptions"`
		UploadThumbnail bool   `json:"uploadThumbnail"`
	} `json:"initializeUploadRequest"`
}

type initVideoUploadResp struct {
	Value struct {
		Video              string `json:"video"`
		UploadUrlsExpireAt int64  `json:"uploadUrlsExpireAt"`
		UploadToken        string `json:"uploadToken"`
		UploadInstructions []struct {
			UploadUrl string `json:"uploadUrl"`
			FirstByte int64  `json:"firstByte"`
			LastByte  int64  `json:"lastByte"`
		} `json:"uploadInstructions"`
		ThumbnailUploadInstruction struct {
			UploadUrl string `json:"uploadUrl"`
		} `json:"thumbnailUploadInstruction"`
	} `json:"value"`
}

type finalizeVideoUploadReq struct {
	FinalizeUploadRequest struct {
		Video           string   `json:"video"`
		UploadToken     string   `json:"uploadToken"`
		UploadedPartIds []string `json:"uploadedPartIds"`
	} `json:"finalizeUploadRequest"`
}

func (vp *VideoPoster) Post(ctx context.Context, pp *post.PublishPost, mediaList []*media.Media) error {
	if err := vp.Validate(ctx, pp, mediaList); err != nil {
		return err
	}
	m := mediaList[0]
	// Step 1: Initialize video upload
	initReqBody := initVideoUploadReq{}
	initReqBody.InitializeUploadRequest.Owner = vp.authorURN
	initReqBody.InitializeUploadRequest.FileSizeBytes = int64(len(m.Data))
	initReqBody.InitializeUploadRequest.UploadCaptions = false
	initReqBody.InitializeUploadRequest.UploadThumbnail = (m.Thumbnail != nil)

	initJSON, err := json.Marshal(initReqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal init video upload request: %w", err)
	}

	initReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linkedin.com/rest/videos?action=initializeUpload", bytes.NewBuffer(initJSON))
	if err != nil {
		return fmt.Errorf("failed to create init video request: %w", err)
	}
	setHeaders(initReq, vp.secrets.AccessToken)

	initResp, err := vp.httpClient.Do(initReq)
	if err != nil {
		return fmt.Errorf("failed to send init video request: %w", err)
	}
	defer initResp.Body.Close()

	if initResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(initResp.Body)
		return fmt.Errorf("initialize video upload failed (%d): %s", initResp.StatusCode, string(body))
	}

	var initRes initVideoUploadResp
	if err := json.NewDecoder(initResp.Body).Decode(&initRes); err != nil {
		return fmt.Errorf("failed to decode init video upload response: %w", err)
	}

	videoURN := initRes.Value.Video
	if videoURN == "" || len(initRes.Value.UploadInstructions) == 0 {
		return errors.New("invalid init response: missing video URN or upload instructions")
	}

	// Step 2: Upload video chunks
	uploadedPartIds := make([]string, 0, len(initRes.Value.UploadInstructions))
	for _, instr := range initRes.Value.UploadInstructions {
		chunk := m.Data[instr.FirstByte : instr.LastByte+1]

		uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPut, instr.UploadUrl, bytes.NewBuffer(chunk))
		if err != nil {
			return fmt.Errorf("failed to create chunk upload request: %w", err)
		}

		setBinaryHeaders(uploadReq, vp.secrets.AccessToken)

		uploadResp, err := vp.httpClient.Do(uploadReq)
		if err != nil {
			return fmt.Errorf("failed to upload chunk: %w", err)
		}

		// Get the ETag from response headers - this is our part ID
		etag := uploadResp.Header.Get("etag")
		if etag == "" {
			body, _ := io.ReadAll(uploadResp.Body)
			uploadResp.Body.Close()
			return fmt.Errorf("no etag in upload response: %s", string(body))
		}
		uploadResp.Body.Close()

		if uploadResp.StatusCode != http.StatusOK {
			return fmt.Errorf("chunk upload failed with status: %d", uploadResp.StatusCode)
		}

		uploadedPartIds = append(uploadedPartIds, etag)
	}

	// Step 2b: Upload thumbnail if present
	if m.Thumbnail != nil && initRes.Value.ThumbnailUploadInstruction.UploadUrl != "" {
		thumbReq, err := http.NewRequestWithContext(ctx, http.MethodPut, initRes.Value.ThumbnailUploadInstruction.UploadUrl, bytes.NewBuffer(m.Thumbnail.Data))
		if err != nil {
			return fmt.Errorf("failed to create thumbnail upload request: %w", err)
		}

		// Set headers as per documentation
		thumbReq.Header.Set("Content-Type", "application/octet-stream")
		thumbReq.Header.Set("media-type-family", "STILLIMAGE")
		thumbReq.Header.Set("Authorization", "Bearer "+vp.secrets.AccessToken)

		thumbResp, err := vp.httpClient.Do(thumbReq)
		if err != nil {
			return fmt.Errorf("failed to upload thumbnail: %w", err)
		}
		defer thumbResp.Body.Close()

		if thumbResp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(thumbResp.Body)
			return fmt.Errorf("thumbnail upload failed (%d): %s", thumbResp.StatusCode, string(body))
		}
	}

	// Step 3: Finalize video upload
	finalizeReqBody := finalizeVideoUploadReq{}
	finalizeReqBody.FinalizeUploadRequest.Video = videoURN
	finalizeReqBody.FinalizeUploadRequest.UploadToken = initRes.Value.UploadToken
	finalizeReqBody.FinalizeUploadRequest.UploadedPartIds = uploadedPartIds

	finalizeJSON, err := json.Marshal(finalizeReqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal finalize request: %w", err)
	}

	finalizeReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linkedin.com/rest/videos?action=finalizeUpload", bytes.NewBuffer(finalizeJSON))
	if err != nil {
		return fmt.Errorf("failed to create finalize request: %w", err)
	}
	setHeaders(finalizeReq, vp.secrets.AccessToken)

	finalizeResp, err := vp.httpClient.Do(finalizeReq)
	if err != nil {
		return fmt.Errorf("failed to finalize video upload: %w", err)
	}
	defer finalizeResp.Body.Close()

	if finalizeResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(finalizeResp.Body)
		return fmt.Errorf("video finalize failed (%d): %s", finalizeResp.StatusCode, string(body))
	}

	// Step 4: Create the post with the video
	postBody := map[string]interface{}{
		"author":     vp.authorURN,
		"commentary": pp.TextContent,
		"visibility": "PUBLIC",
		"distribution": map[string]interface{}{
			"feedDistribution":               "MAIN_FEED",
			"targetEntities":                 []string{},
			"thirdPartyDistributionChannels": []string{},
		},
		"content": map[string]interface{}{
			"media": map[string]interface{}{
				"id": videoURN,
			},
		},
		"lifecycleState":            "PUBLISHED",
		"isReshareDisabledByAuthor": false,
	}

	postJSON, err := json.Marshal(postBody)
	if err != nil {
		return fmt.Errorf("failed to marshal post request: %w", err)
	}

	postReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linkedin.com/rest/posts", bytes.NewBuffer(postJSON))
	if err != nil {
		return fmt.Errorf("failed to create post request: %w", err)
	}
	setHeaders(postReq, vp.secrets.AccessToken)

	postResp, err := vp.httpClient.Do(postReq)
	if err != nil {
		return fmt.Errorf("failed to send post request: %w", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusOK && postResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(postResp.Body)
		return fmt.Errorf("create post failed (%d): %s", postResp.StatusCode, string(body))
	}

	return nil
}
