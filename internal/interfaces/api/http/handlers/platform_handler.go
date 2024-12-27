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
		e.WriteBusinessError(w, err, nil)
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
	APIKey           string `json:"api_key"`
}

// AddAPIKey godoc
// @Summary Add an API key to a social network
// @Description Add an API key to a social network
// @Tags platforms
// @Accept json
// @Produce json
// @Param api_key body addAPIKeyRequest true "API key request"
// @Success 200
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /platforms/api-key [post]
func (h *PlatformHandler) AddAPIKey(w http.ResponseWriter, r *http.Request) {
	var req addAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e.WriteBusinessError(w, e.NewValidationError("Invalid request payload", nil), nil)
		return
	}

	if req.SocialPlatformID == "" {
		e.WriteHttpError(w, e.NewValidationError("Social network id is required", map[string]string{
			"social_network_id": "required",
		}))
		return
	}

	if req.APIKey == "" {
		e.WriteHttpError(w, e.NewValidationError("API key is required", map[string]string{
			"api_key": "required",
		}))
		return
	}

	projectID := r.PathValue("project_id")

	err := h.Service.AddAPIKey(r.Context(), projectID, req.SocialPlatformID, req.APIKey)
	if err != nil {
		e.WriteBusinessError(w, err, mapPlatformErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}
