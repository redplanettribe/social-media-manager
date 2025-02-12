package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/domain/publisher"
	e "github.com/pedrodcsjostrom/opencm/internal/utils/errors"
)

type PublisherHandler struct {
	Service publisher.Service
}

func NewPlatformHandler(service publisher.Service) *PublisherHandler {
	return &PublisherHandler{Service: service}
}

// GetAvailableSocialNetworks godoc
// @Summary Get available social networks
// @Description Get available social networks
// @Tags publishers
// @Accept json
// @Produce json
// @Success 200 {object} []publisher.Platform
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /publishers [get]
func (h *PublisherHandler) GetAvailableSocialNetworks(w http.ResponseWriter, r *http.Request) {
	publishers, err := h.Service.GetAvailableSocialNetworks(r.Context())
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(publishers)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

// PublishPostToSocialNetwork godoc
// @Summary Publish post to social network
// @Description Publish post to social network
// @Tags publishers
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param post_id path string true "Post ID"
// @Param social_network_id path string true "Social Network ID"
// @Success 200
// @Failure 400 {object} errors.APIError "Bad request"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /publishers/{project_id}/{post_id}/{platform_id} [post]
func (h *PublisherHandler) PublishPostToSocialNetwork(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id":  "required",
		"post_id":     "required",
		"platform_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}

	postID := r.PathValue("post_id")
	projectID := r.PathValue("project_id")
	socialNetworkID := r.PathValue("platform_id")

	err := h.Service.PublishPostToSocialNetwork(r.Context(), projectID, postID, socialNetworkID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PublisherHandler) PublishPostToAssignedSocialNetworks(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
		"post_id":    "required",
	}
	if !requirePathParams(w, params) {
		return
	}

	postID := r.PathValue("post_id")
	projectID := r.PathValue("project_id")

	err := h.Service.PublishPostToAssignedSocialNetworks(r.Context(), projectID, postID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type authenticationRequest struct {
	Params map[string]interface{} `json:"params"`
}

func (r authenticationRequest) Validate() map[string]string {
	errors := make(map[string]string)
	if r.Params == nil {
		errors["params"] = "required"
	}
	return errors
}

// Authenticate godoc
// @Summary Authenticate
// @Description Authenticate
// @Tags publishers
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param platform_id path string true "Platform ID"
// @Param user_id path string true "User ID"
// @Param authentication body authenticationRequest true "Authentication request"
// @Success 200
// @Failure 400 {object} errors.APIError "Bad request"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /publishers/{project_id}/{platform_id}/{user_id}/authenticate [post]
func (h *PublisherHandler) Authenticate(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id":  "required",
		"platform_id": "required",
		"user_id":     "required",
	}
	if !requirePathParams(w, params) {
		return
	}

	platformID := r.PathValue("platform_id")
	userID := r.PathValue("user_id")
	projectID := r.PathValue("project_id")

	req, ok := validateRequestBody[authenticationRequest](w, r)
	if !ok {
		return
	}

	err := h.Service.Authenticate(r.Context(), platformID, projectID, userID, req.Params)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ValidatePostForAllAssignedSocialNetworks godoc
// @Summary Validate post for all assigned social networks
// @Description Validate post for all assigned social networks
// @Tags publishers
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param post_id path string true "Post ID"
// @Success 200
// @Failure 400 {object} errors.APIError "Bad request"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /publishers/{project_id}/{post_id}/validate [get]
func (h *PublisherHandler) ValidatePostForAssignedSocialNetworks(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
		"post_id":    "required",
	}
	if !requirePathParams(w, params) {
		return
	}

	projectID := r.PathValue("project_id")
	postID := r.PathValue("post_id")

	err := h.Service.ValidatePostForAssignedSocialNetworks(r.Context(), projectID, postID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ValidatePostForSocialNetwork godoc
// @Summary Validate post for social network
// @Description Validate post for social network
// @Tags publishers
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param post_id path string true "Post ID"
// @Param social_network_id path string true "Social Network ID"
// @Success 200
// @Failure 400 {object} errors.APIError "Bad request"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /publishers/{project_id}/{post_id}/{platform_id}/validate [get]
func (h *PublisherHandler) ValidatePostForSocialNetwork(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id":  "required",
		"post_id":     "required",
		"platform_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}

	projectID := r.PathValue("project_id")
	postID := r.PathValue("post_id")
	platformID := r.PathValue("platform_id")

	err := h.Service.ValidatePostForSocialNetwork(r.Context(), projectID, postID, platformID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetPublishPostInfo godoc
// @Summary Get publish post info
// @Description Get publish post info
// @Tags publishers
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param post_id path string true "Post ID"
// @Param social_network_id path string true "Social Network ID"
// @Success 200 {object} publisher.PublishPostInfo
// @Failure 400 {object} errors.APIError "Bad request"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /publishers/{project_id}/{post_id}/{platform_id}/info [get]
func (h *PublisherHandler) GetPublishPostInfo(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id":  "required",
		"post_id":     "required",
		"platform_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}

	projectID := r.PathValue("project_id")
	postID := r.PathValue("post_id")
	platformID := r.PathValue("platform_id")

	info, err := h.Service.GetPublishPostInfo(r.Context(), projectID, postID, platformID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(info)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}
