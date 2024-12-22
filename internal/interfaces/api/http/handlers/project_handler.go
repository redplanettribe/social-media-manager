package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/domain/project"
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
// @Failure 400 {string} string "Invalid request payload"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Security ApiKeyAuth
// @Router /projects [post]
func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	p, err := h.Service.CreateProject(ctx, req.Name, req.Description)
	if err != nil {
		statusCode, message := MapErrorToHTTP(err)
		http.Error(w, message, statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

// ListProjects godoc
// @Summary List all projects
// @Description List all projects that the user is a member of
// @Tags projects
// @Accept json
// @Produce json
// @Success 200 {array} project.Project
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal server error"
// @Security ApiKeyAuth
// @Router /projects [get]
func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projects, err := h.Service.ListProjects(ctx)
	if err != nil {
		statusCode, message := MapErrorToHTTP(err)
		http.Error(w, message, statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(projects)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
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
// @Failure 400 {string} string "Invalid request payload"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Project not found"
// @Failure 500 {string} string "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id} [get]
func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := r.PathValue("project_id")
	if projectID == "" {
		http.Error(w, "project id not found in query params", http.StatusBadRequest)
		return
	}

	p, err := h.Service.GetProject(ctx, projectID)
	if err != nil {
		statusCode, message := MapErrorToHTTP(err)
		http.Error(w, message, statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
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
// @Param user_id path string true "User ID"
// @Success 204 {string} string "No content"
// @Failure 400 {string} string "Invalid request payload"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Project not found"
// @Failure 500 {string} string "Internal server error"
// @Security ApiKeyAuth
// @Router /projects/{project_id}/add [post]
func (h *ProjectHandler) AddUserToProject(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	projectID := r.PathValue("project_id")
	if projectID == "" {
		http.Error(w, "project id not found in query params", http.StatusBadRequest)
		return
	}

	var req addUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err := h.Service.AddUserToProject(ctx, projectID, req.Email)
	if err != nil {
		statusCode, message := MapErrorToHTTP(err)
		http.Error(w, message, statusCode)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
