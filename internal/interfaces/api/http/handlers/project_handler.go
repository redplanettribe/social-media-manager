package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/domain/project"
	e "github.com/pedrodcsjostrom/opencm/internal/utils/errors"
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
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e.WriteBusinessError(w, e.NewValidationError("Invalid request payload", nil), nil)
		return
	}

	if req.Name == "" {
		e.WriteBusinessError(w, e.NewValidationError("Name is required", map[string]string{
			"name": "required",
		}), nil)
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
	ctx := r.Context()
	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteBusinessError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}), nil)
		return
	}

	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e.WriteBusinessError(w, e.NewValidationError("Invalid request payload", nil), nil)
		return
	}

	if req.Name == "" {
		e.WriteBusinessError(w, e.NewValidationError("Name is required", map[string]string{
			"name": "required",
		}), nil)
		return
	}

	p, err := h.Service.UpdateProject(ctx, projectID, req.Name, req.Description)
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
	ctx := r.Context()
	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteBusinessError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}), nil)
		return
	}

	p, err := h.Service.GetProject(ctx, projectID)
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
	ctx := r.Context()
	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteBusinessError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}), nil)
		return
	}

	var req addUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e.WriteBusinessError(w, e.NewValidationError("Invalid request payload", nil), nil)
		return
	}
	if req.Email == "" {
		e.WriteBusinessError(w, e.NewValidationError("Email is required", map[string]string{
			"email": "required",
		}), nil)
		return
	}

	err := h.Service.AddUserToProject(ctx, projectID, req.Email)
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
// @Router /projects/{project_id}/enable-social-platform/{social_platform_id} [post]
func (h *ProjectHandler) EnableSocialPlatform(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteBusinessError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}), nil)
		return
	}

	socialPlatformID := r.PathValue("platform_id")
	if socialPlatformID == "" {
		e.WriteBusinessError(w, e.NewValidationError("Social platform id is required", map[string]string{
			"platform_id": "required",
		}), nil)
		return
	}

	err := h.Service.EnableSocialPlatform(ctx, projectID, socialPlatformID)
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
	ctx := r.Context()
	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteBusinessError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}), nil)
		return
	}

	platforms, err := h.Service.GetEnabledSocialPlatforms(ctx, projectID)
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

type setTimeZoneRequest struct {
	TimeZone string `json:"time_zone"`
}

// SetTimeZone godoc
// @Summary Set the time zone for a project
// @Description Set the time zone for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param project_id path string true "Project ID"
// @Param time_zone body setTimeZoneRequest true "Time zone request"
// @Success 204 {string} string "No content"
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 410 {object} errors.APIError "Project not found"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/time-zone [patch]
func (h *ProjectHandler) SetTimeZone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req setTimeZoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e.WriteHttpError(w, e.NewValidationError("Invalid request payload", nil))
		return
	}
	if req.TimeZone == "" {
		e.WriteHttpError(w, e.NewValidationError("Time zone is required", map[string]string{
			"time_zone": "required",
		}))
		return
	}

	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteHttpError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}))
		return
	}

	err := h.Service.SetTimeZone(ctx, projectID, req.TimeZone)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type addTimeSlotRequest struct {
	DayOfWeek int `json:"day_of_week"` // time.Weekday
	Hour      int `json:"hour"`
	Minute    int `json:"minute"`
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
// @Router /projects/{project_id}/time-slots [patch]
func (h *ProjectHandler) AddTimeSlot(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteHttpError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}))
		return
	}

	var req addTimeSlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		e.WriteHttpError(w, e.NewValidationError("Invalid request payload", nil))
		return
	}

	err := h.Service.AddTimeSlot(ctx, projectID, time.Weekday(req.DayOfWeek), req.Hour, req.Minute)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	ctx := r.Context()
	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteHttpError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}))
		return
	}

	userID := r.PathValue("user_id")
	if userID == "" {
		e.WriteHttpError(w, e.NewValidationError("User id is required", map[string]string{
			"user_id": "required",
		}))
		return
	}

	err := h.Service.SetDefaultUser(ctx, projectID, userID)
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
	ctx := r.Context()
	projectID := r.PathValue("project_id")
	if projectID == "" {
		e.WriteHttpError(w, e.NewValidationError("Project id is required", map[string]string{
			"project_id": "required",
		}))
		return
	}

	platformID := r.PathValue("platform_id")
	if platformID == "" {
		e.WriteHttpError(w, e.NewValidationError("Platform id is required", map[string]string{
			"platform_id": "required",
		}))
		return
	}

	info, err := h.Service.GetDefaultUserPlatformInfo(ctx, projectID, platformID)
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
