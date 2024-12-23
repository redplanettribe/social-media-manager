package errors

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/domain/project"
	"github.com/pedrodcsjostrom/opencm/internal/domain/user"
)

// WriteError writes a standardized error response
func WriteError(w http.ResponseWriter, err error) {
	var apiError *APIError
	if !errors.As(err, &apiError) {
		apiError = mapErrorToAPIError(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiError.Status)
	json.NewEncoder(w).Encode(apiError)
}

// mapErrorToAPIError maps domain errors to APIErrors
func mapErrorToAPIError(err error) *APIError {
	switch {
	case matchError(
		err,
		user.ErrExistingUser,
		project.ErrProjectExists,
	):
		return &APIError{
			Status:  http.StatusConflict,
			Code:    ErrCodeConflict,
			Message: err.Error(),
		}
	case matchError(
		err,
		user.ErrUserNotFound,
		project.ErrProjectNotFound,
	):
		return &APIError{
			Status:  http.StatusGone,
			Code:    ErrCodeNotFound,
			Message: err.Error(),
		}
	case matchError(
		err,
		user.ErrInvalidPassword,
	):
		return &APIError{
			Status:  http.StatusForbidden,
			Code:    ErrCodeForbidden,
			Message: "Invalid password",
		}

	default:
		return &APIError{
			Status:  http.StatusInternalServerError,
			Code:    ErrCodeInternal,
			Message: "Internal server error",
		}
	}
}

// matchError checks if err matches any of the provided errors
func matchError(err error, errs ...error) bool {
	for _, e := range errs {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}
