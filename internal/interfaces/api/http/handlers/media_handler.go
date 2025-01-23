package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	e "github.com/pedrodcsjostrom/opencm/internal/utils/errors"
)

type MediaHandler struct {
	Service media.Service
}

func NewMediaHandler(service media.Service) *MediaHandler {
	return &MediaHandler{
		Service: service,
	}
}

// UploadMedia godoc
// @Summary Upload media
// @Description Upload media
// @Tags media
// @Accept mpfd
// @Param project_id path string true "Project ID"
// @Param post_id path string true "Post ID"
// @Param file formData file true "File"
// @Param alt_text formData string false "Alt text"
// @Success 201 {object} media.DownloadMetaData
// @Failure 400 {object} errors.APIError
// @Failure 401 {object} errors.APIError
// @Failure 403 {object} errors.APIError
// @Failure 404 {object} errors.APIError
// @Failure 500 {object} errors.APIError
// @Router /media/{project_id}/{post_id} [post]
func (h *MediaHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteHttpError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}))
		return
	}
	postID := r.PathValue("post_id")
	if postID == "" {
		e.WriteHttpError(w, e.NewValidationError("Post id is required", map[string]string{
			"post_id": "required",
		}))
		return
	}

	altText := r.FormValue("alt_text")

	file, header, err := r.FormFile("file")
	if err != nil {
		e.WriteHttpError(w, e.NewValidationError("Invalid file", map[string]string{
			"file": "invalid",
		}))
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	downloadMetadata, err := h.Service.UploadMedia(r.Context(), projectID, postID, header.Filename, altText, data)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(downloadMetadata)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

// GetMediaFile godoc
// @Summary Get media file
// @Description Get media file. This endpoint shouldn't be used. Use the frontend to get the media file directly from the bucket.
// @Tags media
// @Produce octet-stream
// @Param project_id path string true "Project ID"
// @Param post_id path string true "Post ID"
// @Param file_name path string true "File name"
// @Success 200 {string} string
// @Failure 400 {object} errors.APIError
// @Failure 401 {object} errors.APIError
// @Failure 403 {object} errors.APIError
// @Failure 404 {object} errors.APIError
// @Failure 500 {object} errors.APIError
// @Router /media/{project_id}/{post_id}/{file_name} [get]
func (h *MediaHandler) GetMediaFile(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteHttpError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}))
		return
	}
	postID := r.PathValue("post_id")
	if postID == "" {
		e.WriteHttpError(w, e.NewValidationError("Post id is required", map[string]string{
			"post_id": "required",
		}))
		return
	}

	filename := r.PathValue("file_name")
	if filename == "" {
		e.WriteHttpError(w, e.NewValidationError("Filename is required", map[string]string{
			"file_name": "required",
		}))
		return
	}

	m, err := h.Service.GetMediaFile(r.Context(), projectID, postID, filename)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", fmt.Sprintf("%s/%s", m.Type, m.Format))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(m.Data)))
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(m.Data)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to write response"))
	}
}

type linkMediaToPublishPostRequest struct {
	MediaID    string `json:"media_id"`
	PlatformID string `json:"platform_id"`
	PostID     string `json:"post_id"`
}

// LinkMediaToPublishPost godoc
// @Summary Link media to publish post
// @Description Link media to publish post
// @Tags media
// @Accept json
// @Param project_id path string true "Project ID"
// @Param post_id path string true "Post ID"
// @Param platform_id path string true "Platform ID"
// @Param media_id body linkMediaToPublishPostRequest true "Media ID"
// @Success 204
// @Failure 400 {object} errors.APIError
// @Failure 401 {object} errors.APIError
// @Failure 403 {object} errors.APIError
// @Failure 404 {object} errors.APIError
// @Failure 500 {object} errors.APIError
// @Router /media/{project_id}/link-to-platform [post]
func (h *MediaHandler) LinkMediaToPublishPost(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteHttpError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}))
		return
	}

	var req linkMediaToPublishPostRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		e.WriteHttpError(w, e.NewValidationError("Invalid request", nil))
		return
	}

	mediaID, platformID, postID := req.MediaID, req.PlatformID, req.PostID
	if mediaID == "" || platformID == "" || postID == "" {
		e.WriteHttpError(w, e.NewValidationError("Media id, platform id and post id are required", map[string]string{
			"media_id":    "required",
			"platform_id": "required",
			"post_id":     "required",
		}))
		return
	}

	err = h.Service.LinkMediaToPublishPost(r.Context(), projectID, postID, mediaID, platformID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDownloadMetaData godoc
// @Summary Get download metadata
// @Description Get download metadata
// @Tags media
// @Produce json
// @Param project_id path string true "Project ID"
// @Param post_id path string true "Post ID"
// @Param file_name path string true "File name"
// @Success 200 {object} media.DownloadMetaData
// @Failure 400 {object} errors.APIError
// @Failure 401 {object} errors.APIError
// @Failure 403 {object} errors.APIError
// @Failure 404 {object} errors.APIError
// @Failure 500 {object} errors.APIError
// @Router /media/{project_id}/{post_id}/{file_name}/meta [get]
func (h *MediaHandler) GetDownloadMetaData(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteHttpError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}))
		return
	}
	postID := r.PathValue("post_id")
	if postID == "" {
		e.WriteHttpError(w, e.NewValidationError("Post id is required", map[string]string{
			"post_id": "required",
		}))
		return
	}
	filename := r.PathValue("file_name")
	if filename == "" {
		e.WriteHttpError(w, e.NewValidationError("Filename is required", map[string]string{
			"file_name": "required",
		}))
		return
	}

	downloadMetadata, err := h.Service.GetDownloadMetaData(r.Context(), projectID, postID, filename)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(downloadMetadata)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

func (h *MediaHandler) GetDownloadMetadataDataForPost(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteHttpError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}))
		return
	}
	postID := r.PathValue("post_id")
	if postID == "" {
		e.WriteHttpError(w, e.NewValidationError("Post id is required", map[string]string{
			"post_id": "required",
		}))
		return
	}

	downloadMetadata, err := h.Service.GetDownloadMetadataDataForPost(r.Context(), projectID, postID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(downloadMetadata)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}
