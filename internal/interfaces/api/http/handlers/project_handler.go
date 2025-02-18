package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/redplanettribe/social-media-manager/internal/domain/project"
	e "github.com/redplanettribe/social-media-manager/internal/utils/errors"
)

type ProjectHandler struct {
	Service project.Service
}

func NewProjectHandler(service project.Service) *ProjectHandler {
	return &ProjectHandler{Service: service}
}

type createProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r createProjectRequest) Validate() map[string]string {
	errors := make(map[string]string)
	if r.Name == "" {
		errors["name"] = "Name is required"
	}
	return errors
}

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new project with the given name and description
// @Tags projects
// @Accept json
// @Produce json
// @Param project body createProjectRequest true "Project creation request"
// @Success 201 {object} project.Project
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 409 {object} errors.APIError "Project already exists"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects [post]
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	req, ok := validateRequestBody[createProjectRequest](w, r)
	if !ok {
		return
	}
	p, err := h.Service.CreateProject(r.Context(), req.Name, req.Description)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

// UpdateProject godoc
// @Summary Update a project
// @Description Update a project with the given ID
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param project body createProjectRequest true "Project update request"
// @Success 200 {object} project.Project
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id} [patch]
func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	req, ok := validateRequestBody[createProjectRequest](w, r)
	if !ok {
		return
	}
	params := map[string]string{
		"project_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")

	p, err := h.Service.UpdateProject(r.Context(), projectID, req.Name, req.Description)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

// DeleteProject godoc
// @Summary Delete a project
// @Description Delete a project with the given ID
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Success 204 {string} string "No content"
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id} [delete]
func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")

	err := h.Service.DeleteProject(r.Context(), projectID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListProjects godoc
// @Summary List all projects
// @Description List all projects that the user is a member of
// @Tags projects
// @Accept json
// @Produce json
// @Success 200 {array} project.Project
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects [get]
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projects, err := h.Service.ListProjects(ctx)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(projects)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

// GetProject godoc
// @Summary Get a project
// @Description Get a project by its ID
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Success 200 {object} project.Project
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id} [get]
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")

	p, err := h.Service.GetProject(r.Context(), projectID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

type addUserRequest struct {
	Email string `json:"email"`
}

func (r addUserRequest) Validate() map[string]string {
	errors := make(map[string]string)
	if r.Email == "" {
		errors["email"] = "Email is required"
	}
	return errors
}

// AddUserToProject godoc
// @Summary Add a user to a project
// @Description Add a user to a project by their email
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param user body addUserRequest true "User addition request"
// @Success 204 {string} string "No content"
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 409 {object} errors.APIError "User already exists"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/add-user [post]
func (h *ProjectHandler) AddUserToProject(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")

	req, ok := validateRequestBody[addUserRequest](w, r)
	if !ok {
		return
	}

	err := h.Service.AddUserToProject(r.Context(), projectID, req.Email)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUserRoles godoc
// @Summary Get user roles
// @Description Get the roles of a user in a project
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param user_id path string true "User ID"
// @Success 200 {array} string
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/user-roles/{user_id} [get]
func (h *ProjectHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
		"user_id":    "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")
	userID := r.PathValue("user_id")
	roles, err := h.Service.GetUserRoles(r.Context(), userID, projectID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(roles)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

// AddRoleToUser godoc
// @Summary Add a role to a user
// @Description Add a role to a user in a project
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param user_id path string true "User ID"
// @Param role_id path string true "Role ID"
// @Success 204 {string} string "No content"
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/add-role/{user_id}/{role_id} [post]
func (h *ProjectHandler) AddRoleToUser(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
		"user_id":    "required",
		"role_id":    "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")
	userID := r.PathValue("user_id")
	roleIDStr := r.PathValue("role_id")

	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		e.WriteBusinessError(w, e.NewValidationError("Invalid role id", map[string]string{
			"role_id": "invalid",
		}), nil)
		return
	}

	err = h.Service.AddUserRole(r.Context(), projectID, userID, roleID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveRoleFromUser godoc
// @Summary Remove a role from a user
// @Description Remove a role from a user in a project
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param user_id path string true "User ID"
// @Param role_id path string true "Role ID"
// @Success 204 {string} string "No content"
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/remove-role/{user_id}/{role_id} [delete]
func (h *ProjectHandler) RemoveRoleFromUser(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
		"user_id":    "required",
		"role_id":    "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")
	userID := r.PathValue("user_id")
	roleIDStr := r.PathValue("role_id")

	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		e.WriteBusinessError(w, e.NewValidationError("Invalid role id", map[string]string{
			"role_id": "invalid",
		}), nil)
		return
	}

	err = h.Service.RemoveUserRole(r.Context(), projectID, userID, roleID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveUserFromProject godoc
// @Summary Remove a user from a project
// @Description Remove a user from a project by their email
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param user_id path string true "User ID"
// @Success 204 {string} string "No content"
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/remove-user/{user_id} [delete]
func (h *ProjectHandler) RemoveUserFromProject(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
		"user_id":    "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")
	userID := r.PathValue("user_id")

	err := h.Service.RemoveUserFromProject(r.Context(), projectID, userID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EnableSocialPlatform godoc
// @Summary Enable a social platform
// @Description Enable a social platform for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param social_platform_id path string true "Social Platform ID"
// @Success 204 {string} string "No content"
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 409 {object} errors.APIError "User already exists"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/enable-social-platform/{platform_id} [post]
func (h *ProjectHandler) EnableSocialPlatform(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id":  "required",
		"platform_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")
	socialPlatformID := r.PathValue("platform_id")

	err := h.Service.EnableSocialPlatform(r.Context(), projectID, socialPlatformID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DisableSocialPlatform godoc
// @Summary Disable a social platform
// @Description Disable a social platform for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param social_platform_id path string true "Social Platform ID"
// @Success 204 {string} string "No content"
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 403 {object} errors.APIError "Forbidden"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/disable-social-platform/{platform_id} [delete]
func (h *ProjectHandler) DisableSocialPlatform(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id":  "required",
		"platform_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")
	socialPlatformID := r.PathValue("platform_id")

	err := h.Service.DisableSocialPlatform(r.Context(), projectID, socialPlatformID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetEnabledSocialPlatforms godoc
// @Summary Get enabled social platforms
// @Description Get the social platforms enabled for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Success 200 {array} project.SocialPlatform
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/social-platforms [get]
func (h *ProjectHandler) GetEnabledSocialPlatforms(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")
	platforms, err := h.Service.GetEnabledSocialPlatforms(r.Context(), projectID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(platforms)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

type addTimeSlotRequest struct {
	DayOfWeek int `json:"day_of_week"` // time.Weekday
	Hour      int `json:"hour"`
	Minute    int `json:"minute"`
}

func (r addTimeSlotRequest) Validate() map[string]string {
	errors := make(map[string]string)
	if r.DayOfWeek < 0 || r.DayOfWeek > 6 {
		errors["day_of_week"] = "Invalid day of week"
	}
	if r.Hour < 0 || r.Hour > 23 {
		errors["hour"] = "Invalid hour"
	}
	if r.Minute < 0 || r.Minute > 59 {
		errors["minute"] = "Invalid minute"
	}
	return errors
}

// AddTimeSlot godoc
// @Summary Add a time slot to a project
// @Description Add a time slot to a project
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param time_slot body addTimeSlotRequest true "Time slot request"
// @Success 204 {string} string "No content"
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/add-time-slot [patch]
func (h *ProjectHandler) AddTimeSlot(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")

	req, ok := validateRequestBody[addTimeSlotRequest](w, r)
	if !ok {
		return
	}

	err := h.Service.AddTimeSlot(r.Context(), projectID, time.Weekday(req.DayOfWeek), req.Hour, req.Minute)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveTimeSlot godoc
// @Summary Remove a time slot from a project
// @Description Remove a time slot from a project
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param time_slot body addTimeSlotRequest true "Time slot request"
// @Success 204 {object} nil "No Content"
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/remove-time-slot [patch]
func (h *ProjectHandler) RemoveTimeSlot(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")

	req, ok := validateRequestBody[addTimeSlotRequest](w, r)
	if !ok {
		return
	}

	err := h.Service.RemoveTimeSlot(r.Context(), projectID, time.Weekday(req.DayOfWeek), req.Hour, req.Minute)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetProjectSchedule godoc
// @Summary Get the project schedule
// @Description Get the project schedule for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Success 200 {object} project.WeeklyPostSchedule
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/schedule [get]
func (h *ProjectHandler) GetProjectSchedule(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")

	schedule, err := h.Service.GetProjectSchedule(r.Context(), projectID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(schedule)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

// SetDefaultUser godoc
// @Summary Set the default user for a project
// @Description Set the default user for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param user_id path string true "User ID"
// @Success 204 {string} string "No content"
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/default-user/{user_id} [patch]
func (h *ProjectHandler) SetDefaultUser(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id": "required",
		"user_id":    "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")
	userID := r.PathValue("user_id")

	err := h.Service.SetDefaultUser(r.Context(), projectID, userID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDefaultUserPlatformInfo godoc
// @Summary Get the default user platform info
// @Description Get the default user platform info for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param platform_id path string true "Platform ID"
// @Success 200 {object} project.UserPlatformInfo
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/default-user-platform-info/{platform_id} [get]
func (h *ProjectHandler) GetDefaultUserPlatformInfo(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"project_id":  "required",
		"platform_id": "required",
	}
	if !requirePathParams(w, params) {
		return
	}
	projectID := r.PathValue("project_id")
	platformID := r.PathValue("platform_id")

	info, err := h.Service.GetDefaultUserPlatformInfo(r.Context(), projectID, platformID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(info)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}
