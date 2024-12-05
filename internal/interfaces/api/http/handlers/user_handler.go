package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/domain/user"
)

const sessionCookieName = "session_id"

type UserHandler struct {
	Service user.Service
}

func NewUserHandler(service user.Service) *UserHandler {
	return &UserHandler{Service: service}
}

type createUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	u, err := h.Service.CreateUser(ctx, req.Username, req.Password, req.Email)
	if err != nil {
		statusCode, message := MapErrorToHTTP(err)
		http.Error(w, message, statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(u)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}

}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	u, err := h.Service.GetUser(ctx, id)
	if err != nil {
		statusCode, message := MapErrorToHTTP(err)
		http.Error(w, message, statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(u)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Missing email or password", http.StatusBadRequest)
		return
	}

	session, err := h.Service.Login(ctx, req.Email, req.Password)
	if err != nil || session == nil {
		statusCode, message := MapErrorToHTTP(err)
		http.Error(w, message, statusCode)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *UserHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	roles, err := h.Service.GetAllAppRoles(ctx)
	if err != nil {
		statusCode, message := MapErrorToHTTP(err)
		http.Error(w, message, statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(roles)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

type assignRoleRequest struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

func (h *UserHandler) AssignRoleToUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req assignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.RoleID == "" {
		http.Error(w, "Missing user_id or role_id", http.StatusBadRequest)
		return
	}

	err := h.Service.AssignAppRoleToUser(ctx, req.UserID, req.RoleID)
	if err != nil {
		statusCode, message := MapErrorToHTTP(err)
		http.Error(w, message, statusCode)
		return
	}
}

func (h *UserHandler) RemoveRoleFromUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req assignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	userID, roleID := req.UserID, req.RoleID
	if userID == "" || roleID == "" {
		http.Error(w, "Missing user_id or role_id", http.StatusBadRequest)
		return
	}

	err := h.Service.RemoveAppRoleFromUser(ctx, userID, roleID)
	if err != nil {
		statusCode, message := MapErrorToHTTP(err)
		http.Error(w, message, statusCode)
		return
	}
}
