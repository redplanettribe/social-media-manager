package handlers

import (
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/domain/platform"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/domain/project"
	"github.com/pedrodcsjostrom/opencm/internal/domain/user"
	e "github.com/pedrodcsjostrom/opencm/internal/utils/errors"
)

func mapPostErrorToAPIError(err error) *e.APIError {
	switch {

	case e.MatchError(
		err,
		post.ErrPostNotFound,
		post.ErrProjectNotFound,
	):
		return &e.APIError{
			Status:  http.StatusGone,
			Code:    e.ErrCodeNotFound,
			Message: err.Error(),
		}
	case e.MatchError(
		err,
		post.ErrPostScheduledTime,
	):
		return &e.APIError{
			Status:  http.StatusBadRequest,
			Code:    e.ErrCodeBadRequest,
			Message: err.Error(),
		}
	case e.MatchError(
		err,
		post.ErrPublisherNotInProject,
	):
		return &e.APIError{
			Status:  http.StatusForbidden,
			Code:    e.ErrCodeForbidden,
			Message: err.Error(),
		}

	default:
		return &e.APIError{
			Status:  http.StatusInternalServerError,
			Code:    e.ErrCodeInternal,
			Message: err.Error(),
		}
	}
}

func mapPlatformErrorToAPIError(err error) *e.APIError {
	switch {
	case e.MatchError(err, platform.ErrSocialPlatformNotFound):
		return &e.APIError{
			Status:  http.StatusGone,
			Code:    e.ErrCodeNotFound,
			Message: err.Error(),
		}

	default:
		return &e.APIError{
			Status:  http.StatusGone,
			Code:    e.ErrCodeNotFound,
			Message: err.Error(),
		}
	}
}

func mapProjectErrorToAPIError(err error) *e.APIError {
	switch {
	case e.MatchError(
		err,
		project.ErrProjectExists,
		project.ErrSocialPlatformAlreadyEnabled,
	):
		return &e.APIError{
			Status:  http.StatusConflict,
			Code:    e.ErrCodeConflict,
			Message: err.Error(),
		}
	case e.MatchError(
		err,
		project.ErrProjectNotFound,
	):
		return &e.APIError{
			Status:  http.StatusGone,
			Code:    e.ErrCodeNotFound,
			Message: err.Error(),
		}
	case e.MatchError(
		err,
		project.ErrUserAlreadyInProject,
	):
		return &e.APIError{
			Status:  http.StatusConflict,
			Code:    e.ErrCodeConflict,
			Message: err.Error(),
		}
	default:
		return &e.APIError{
			Status:  http.StatusInternalServerError,
			Code:    e.ErrCodeInternal,
			Message: "Internal server error",
		}
	}
}

func mapUserErrorToAPIError(err error) *e.APIError {
	switch {
	case e.MatchError(
		err,
		user.ErrExistingUser,
	):
		return &e.APIError{
			Status:  http.StatusConflict,
			Code:    e.ErrCodeConflict,
			Message: err.Error(),
		}
	case e.MatchError(
		err,
		user.ErrUserNotFound,
	):
		return &e.APIError{
			Status:  http.StatusGone,
			Code:    e.ErrCodeNotFound,
			Message: err.Error(),
		}
	case e.MatchError(
		err,
		user.ErrInvalidPassword,
	):
		return &e.APIError{
			Status:  http.StatusForbidden,
			Code:    e.ErrCodeForbidden,
			Message: "Invalid password",
		}
	default:
		return &e.APIError{
			Status:  http.StatusInternalServerError,
			Code:    e.ErrCodeInternal,
			Message: "Internal server error",
		}
	}
}
