package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/domain/publisher"
	e "github.com/pedrodcsjostrom/opencm/internal/utils/errors"
)

type PlatformHandler struct {
	Service publisher.Service
}

func NewPlatformHandler(service publisher.Service) *PlatformHandler {
	return &PlatformHandler{Service: service}
}

// GetAvailableSocialNetworks godoc
// @Summary Get available social networks
// @Description Get available social networks
// @Tags publishers
// @Accept json
// @Produce json
// @Success 200 {object} []platform.Platform
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /publishers [get]
func (h *PlatformHandler) GetAvailableSocialNetworks(w http.ResponseWriter, r *http.Request) {
	publishers, err := h.Service.GetAvailableSocialNetworks(r.Context())
	if err != nil {
		e.WriteBusinessError(w, err, mapPlatformErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(publishers)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

type addAPIKeyRequest struct {
	SocialPlatformID string `json:"social_platform_id"`
	SecretKey        string `json:"secret_key"`
	SecretValue      string `json:"secret_value"`
}

// AddSecret godoc
// @Summary Add a secret to a social network
// @Description Add a secret to a social network
// @Tags publishers
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param request body addAPIKeyRequest true "Request body"
// @Success 201 {object} string
// @Failure 400 {object} errors.APIError "Bad request"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /publishers/{project_id}/secrets [post]
func (h *PlatformHandler) AddSecret(w http.ResponseWriter, r *http.Request) {
	var req addAPIKeyRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		e.WriteHttpError(w, e.NewValidationError("Invalid request body", nil))
		return
	}

	projectID := r.PathValue("project_id")


	err = h.Service.AddSecret(r.Context(), projectID, req.SocialPlatformID, req.SecretKey, req.SecretValue)
	if err != nil {
		e.WriteBusinessError(w, err, mapPlatformErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
