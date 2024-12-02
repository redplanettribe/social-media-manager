package handlers

import (
	"errors"
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/domain/user"
)

// MapErrorToHTTP maps the given error to an appropriate HTTP status code and message.
func MapErrorToHTTP(err error) (int, string) {
	switch {
	case errors.Is(err, user.ErrExistingUser):
		return http.StatusConflict, "User already exists"
	case errors.Is(err, user.ErrUserNotFound):
		return http.StatusNotFound, "User not found"
	case errors.Is(err, user.ErrInvalidPassword):
		return http.StatusUnauthorized, "Invalid password"
	default:
		return http.StatusInternalServerError, "Internal server error"
	}
}
