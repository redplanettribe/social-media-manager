package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

type MultiImagePoster struct {
	secrets    Secrets
	httpClient *http.Client
	authorURN  string
}

func NewMultiImagePoster(s Secrets) *MultiImagePoster {
	return &MultiImagePoster{
		secrets:    s,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		authorURN:  s.URN,
	}
}

func (mp *MultiImagePoster) Validate(ctx context.Context, pp *post.PublishPost, mediaList []*media.Media) error {
	if pp == nil {
		return errors.New("publish post is nil")
	}
	if mp.secrets.AccessToken == "" {
		return errors.New("user access token is not set")
	}
	if mp.authorURN == "" {
		return errors.New("user URN is not set")
	}
	if len(mediaList) < 2 {
		return errors.New("multi-image post requires at least 2 images")
	}
	if !onlyImages(mediaList) {
		return errors.New("multi-image post only supports images")
	}
	return nil
}

type initUploadRequestMulti struct {
	InitializeUploadRequest struct {
		Owner string `json:"owner"`
	} `json:"initializeUploadRequest"`
}

type initUploadResponseMulti struct {
	Value struct {
		UploadUrl          string `json:"uploadUrl"`
		Image              string `json:"image"`
		UploadUrlExpiresAt int64  `json:"uploadUrlExpiresAt"`
	} `json:"value"`
}

func (mp *MultiImagePoster) Post(ctx context.Context, pp *post.PublishPost, mediaList []*media.Media) error {
	if err := mp.Validate(ctx, pp, mediaList); err != nil {
		return err
	}

	var (
		mu     sync.Mutex
		wg     sync.WaitGroup
		images []map[string]string
		upErr  error
	)

	for i, m := range mediaList {
		wg.Add(1)
		go func(idx int, file *media.Media) {
			defer wg.Done()

			initReqBody := initUploadRequestMulti{}
			initReqBody.InitializeUploadRequest.Owner = mp.authorURN

			bodyBytes, err := json.Marshal(initReqBody)
			if err != nil {
				setError(&mu, &upErr, fmt.Errorf("failed to marshal initialize upload: %w", err))
				return
			}

			initReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linkedin.com/rest/images?action=initializeUpload", bytes.NewBuffer(bodyBytes))
			if err != nil {
				setError(&mu, &upErr, fmt.Errorf("failed to create init request: %w", err))
				return
			}
			setHeaders(initReq, mp.secrets.AccessToken)

			resp, err := mp.httpClient.Do(initReq)
			if err != nil {
				setError(&mu, &upErr, fmt.Errorf("failed init request: %w", err))
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				setError(&mu, &upErr, fmt.Errorf("init upload failed code %d", resp.StatusCode))
				return
			}

			var initResp initUploadResponseMulti
			if err := json.NewDecoder(resp.Body).Decode(&initResp); err != nil {
				setError(&mu, &upErr, fmt.Errorf("failed decode initResp: %w", err))
				return
			}
			if initResp.Value.UploadUrl == "" || initResp.Value.Image == "" {
				setError(&mu, &upErr, errors.New("missing uploadUrl or image URN"))
				return
			}

			uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPut, initResp.Value.UploadUrl, bytes.NewBuffer(file.Data))
			if err != nil {
				setError(&mu, &upErr, fmt.Errorf("failed to create upload req: %w", err))
				return
			}
			setBinaryHeaders(uploadReq, mp.secrets.AccessToken)

			uploadResp, err := mp.httpClient.Do(uploadReq)
			if err != nil {
				setError(&mu, &upErr, fmt.Errorf("failed upload request: %w", err))
				return
			}
			defer uploadResp.Body.Close()

			if uploadResp.StatusCode != http.StatusOK && uploadResp.StatusCode != http.StatusCreated {
				respBody, _ := io.ReadAll(uploadResp.Body)
				setError(&mu, &upErr, fmt.Errorf("upload failed %d: %s", uploadResp.StatusCode, string(respBody)))
				return
			}

			mu.Lock()
			images = append(images, map[string]string{
				"id":      initResp.Value.Image,
				"altText": m.AltText,
			})
			mu.Unlock()
		}(i, m)
	}

	wg.Wait()
	if upErr != nil {
		return upErr
	}

	finalBody := map[string]interface{}{
		"author":     mp.authorURN,
		"commentary": pp.TextContent,
		"visibility": "PUBLIC",
		"distribution": map[string]interface{}{
			"feedDistribution":               "MAIN_FEED",
			"targetEntities":                 []string{},
			"thirdPartyDistributionChannels": []string{},
		},
		"lifecycleState":            "PUBLISHED",
		"isReshareDisabledByAuthor": false,
		"content": map[string]interface{}{
			"multiImage": map[string]interface{}{
				"images": images,
			},
		},
	}

	postPayload, err := json.Marshal(finalBody)
	if err != nil {
		return fmt.Errorf("failed to marshal final multi-image post: %w", err)
	}

	postReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.linkedin.com/rest/posts", bytes.NewBuffer(postPayload))
	if err != nil {
		return fmt.Errorf("failed to create final multi-image post req: %w", err)
	}
	setHeaders(postReq, mp.secrets.AccessToken)

	postResp, err := mp.httpClient.Do(postReq)
	if err != nil {
		return fmt.Errorf("failed final multi-image post req: %w", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusOK && postResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(postResp.Body)
		return fmt.Errorf("create multi-image post failed with %d: %s", postResp.StatusCode, string(body))
	}

	fmt.Println("Multi-image post published successfully to LinkedIn.")
	return nil
}

// Helper to set a single error safely and only once
func setError(mu *sync.Mutex, target *error, err error) {
	mu.Lock()
	defer mu.Unlock()
	if *target == nil {
		*target = err
	}
}

func onlyImages(mediaList []*media.Media) bool {
	for _, m := range mediaList {
		if m.Type != media.MediaTypeImage {
			return false
		}
	}
	return true
}
