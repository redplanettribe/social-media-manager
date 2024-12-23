package errors

import (
	"fmt"
	"net/http"
)

// APIError represents a standardized API error. Implements the error interface.
type APIError struct {
    Status  int    `json:"status"`
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

func (e *APIError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Common error codes
const (
    ErrCodeValidation     = "VALIDATION_ERROR"
    ErrCodeNotFound       = "NOT_FOUND"
    ErrCodeUnauthorized   = "UNAUTHORIZED"
    ErrCodeForbidden      = "FORBIDDEN"
    ErrCodeConflict       = "CONFLICT"
    ErrCodeInternal       = "INTERNAL_ERROR"
    ErrCodeBadRequest     = "BAD_REQUEST"
)

func NewValidationError(message string, details any) *APIError {
    return &APIError{
        Status:  http.StatusBadRequest,
        Code:    ErrCodeValidation,
        Message: message,
        Details: details,
    }
}

func NewNotFoundError(message string) *APIError {
    return &APIError{
        Status:  http.StatusGone,
        Code:    ErrCodeNotFound,
        Message: message,
    }
}

func NewUnauthorizedError(message string) *APIError {
	return &APIError{
		Status:  http.StatusUnauthorized,
		Code:    ErrCodeUnauthorized,
		Message: message,
	}
}

func NewForbiddenError(message string) *APIError {
	return &APIError{
		Status:  http.StatusForbidden,
		Code:    ErrCodeForbidden,
		Message: message,
	}
}

func NewConflictError(message string) *APIError {
	return &APIError{
		Status:  http.StatusConflict,
		Code:    ErrCodeConflict,
		Message: message,
	}
}

func NewInternalError(message string) *APIError {
	return &APIError{
		Status:  http.StatusInternalServerError,
		Code:    ErrCodeInternal,
		Message: message,
	}
}

