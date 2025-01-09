package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/domain/platform"
	e "github.com/pedrodcsjostrom/opencm/internal/utils/errors"
)

type PlatformHandler struct {
	Service platform.Service
}

func NewPlatformHandler(service platform.Service) *PlatformHandler {
	return &PlatformHandler{Service: service}
}

// GetAvailableSocialNetworks godoc
// @Summary Get available social networks
// @Description Get available social networks
// @Tags platforms
// @Accept json
// @Produce json
// @Success 200 {object} []platform.Platform
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /platforms [get]
func (h *PlatformHandler) GetAvailableSocialNetworks(w http.ResponseWriter, r *http.Request) {
	platforms, err := h.Service.GetAvailableSocialNetworks(r.Context())
	if err != nil {
		e.WriteBusinessError(w, err, mapPlatformErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(platforms)
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
// @Tags platforms
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param request body addAPIKeyRequest true "Request body"
// @Success 201 {object} string
// @Failure 400 {object} errors.APIError "Bad request"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /platforms/{project_id}/secrets [post]
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
