package handlers

import (
	"encoding/json"
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

type uploadMediaResponse struct {
	*media.MetaData
}

func (h *MediaHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	media, err := h.Service.UploadMedia(ctx, projectID, postID, header.Filename, data)
	if err != nil {
		e.WriteBusinessError(w, err, mapPostErrorToAPIError)
		return
	}
 
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(uploadMediaResponse{media})
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}