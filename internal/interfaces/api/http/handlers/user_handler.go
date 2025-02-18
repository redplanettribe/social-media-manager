package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/redplanettribe/social-media-manager/internal/domain/user"
	e "github.com/redplanettribe/social-media-manager/internal/utils/errors"
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

func (r createUserRequest) Validate() map[string]string {
	errors := make(map[string]string)
	if r.Username == "" {
		errors["username"] = "Username is required"
	}
	if r.Password == "" {
		errors["password"] = "Password is required"
	}
	if r.Email == "" {
		errors["email"] = "Email is required"
	}
	return errors
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r loginRequest) Validate() map[string]string {
	errors := make(map[string]string)
	if r.Email == "" {
		errors["email"] = "Email is required"
	}
	if r.Password == "" {
		errors["password"] = "Password is required"
	}
	return errors
}

// SignUp godoc
// @Summary Register a new user
// @Description Register a new user with username, password and email
// @Tags users
// @Accept json
// @Produce json
// @Param user body createUserRequest true "User creation request"
// @Success 201 {object} user.UserResponse
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 409 {object} errors.APIError "User already exists"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Router /users [post]
func (h *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, ok := validateRequestBody[createUserRequest](w, r)
	if !ok {
		return
	}

	u, err := h.Service.CreateUser(ctx, req.Username, req.Password, req.Email)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(u)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

// GetUser godoc
// @Summary Get user information
// @Description Get information about the currently authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} user.UserResponse
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /users/me [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u, err := h.Service.GetUser(ctx)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(u)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

// Login godoc
// @Summary Login
// @Description Login with email and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body loginRequest true "Login request"
// @Success 200 {object} user.LoginResponse
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Router /users/login [post]
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, ok := validateRequestBody[loginRequest](w, r)
	if !ok {
		return
	}

	response, err := h.Service.Login(ctx, req.Email, req.Password)
	session := response.Session
	if err != nil || session == nil {
		e.WriteHttpError(w, e.NewUnauthorizedError("Invalid email or password"))
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

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		e.WriteHttpError(w, e.NewInternalError("Failed to encode response"))
	}
}

// Logout godoc
// @Summary Logout
// @Description Logout the currently authenticated user
// @Tags users
// @Success 200
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /users/logout [post]
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionID, err := r.Cookie(sessionCookieName)
	if err != nil {
		e.WriteHttpError(w, e.NewValidationError("Missing session cookie", nil))
		return
	}

	err = h.Service.Logout(ctx, sessionID.Value)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    sessionCookieName,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	})

	w.WriteHeader(http.StatusOK)
}

// GetRoles godoc
// @Summary Get all roles
// @Description Get all application roles
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} user.AppRole
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /users/roles [get]
func (h *UserHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	roles, err := h.Service.GetAllAppRoles(ctx)
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

type assignRoleRequest struct {
	UserID string `json:"user_id"`
	RoleID string `json:"role_id"`
}

func (r assignRoleRequest) Validate() map[string]string {
	errors := make(map[string]string)
	if r.UserID == "" {
		errors["user_id"] = "User ID is required"
	}
	if r.RoleID == "" {
		errors["role_id"] = "Role ID is required"
	}
	return errors
}

// AssignRoleToUser godoc
// @Summary Assign role to user
// @Description Assign an application role to a user
// @Tags users
// @Accept json
// @Produce json
// @Param user body assignRoleRequest true "Assign role request"
// @Success 200
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /users/roles [post]
func (h *UserHandler) AssignRoleToUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, ok := validateRequestBody[assignRoleRequest](w, r)
	if !ok {
		return
	}

	err := h.Service.AssignAppRoleToUser(ctx, req.UserID, req.RoleID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RemoveRoleFromUser godoc
// @Summary Remove role from user
// @Description Remove an application role from a user
// @Tags users
// @Accept json
// @Produce json
// @Param user body assignRoleRequest true "Remove role request"
// @Success 200
// @Failure 400 {object} errors.APIError "Validation error"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Security ApiKeyAuth
// @Router /users/roles [delete]
func (h *UserHandler) RemoveRoleFromUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, ok := validateRequestBody[assignRoleRequest](w, r)
	if !ok {
		return
	}

	err := h.Service.RemoveAppRoleFromUser(ctx, req.UserID, req.RoleID)
	if err != nil {
		e.WriteBusinessError(w, err, mapErrorToAPIError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
