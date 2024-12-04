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
