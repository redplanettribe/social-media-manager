package handlers

import (
	"net/http"

	"github.com/redplanettribe/social-media-manager/internal/domain/media"
	"github.com/redplanettribe/social-media-manager/internal/domain/post"
	"github.com/redplanettribe/social-media-manager/internal/domain/project"
	"github.com/redplanettribe/social-media-manager/internal/domain/publisher"
	"github.com/redplanettribe/social-media-manager/internal/domain/user"
	e "github.com/redplanettribe/social-media-manager/internal/utils/errors"
)

func mapErrorToAPIError(err error) *e.APIError {
	switch {
	// Status 400 Bad Request
	case e.MatchError(err,
		post.ErrPostScheduledTime,
		media.ErrUnsupportedMediaType,
		post.ErrInvalidPostType,
		post.ErrPostNotInQueue,
		media.ErrMediaAlreadyLinkedToPost,
		media.ErrFileAlreadyExists,
		media.ErrInvalidMedia,
		media.ErrPostDoesNotBelongToProject,
		media.ErrMediaNotLinkedToPost,
	):
		return &e.APIError{
			Status:  http.StatusBadRequest,
			Code:    e.ErrCodeBadRequest,
			Message: err.Error(),
		}

	// Status 403 Forbidden
	case e.MatchError(err,
		post.ErrPublisherNotInProject,
		publisher.ErrSocialPlatformNotEnabledForProject,
		media.ErrPostDoesNotBelongToProject,
		media.ErrMediaDoesNotBelongToPost,
		media.ErrPlatformNotEnabledForProject,
		media.ErrPostNotLinkedToPlatform,
		media.ErrMediaAlreadyLinkedToPost,
		user.ErrInvalidPassword,
		post.ErrPostNotInProject,
		project.ErrInsufficientPermissions,
	):
		return &e.APIError{
			Status:  http.StatusForbidden,
			Code:    e.ErrCodeForbidden,
			Message: err.Error(),
		}

	// Status 409 Conflict
	case e.MatchError(err,
		post.ErrPostAlreadyInQueue,
		post.ErrPostAlreadyPublished,
		project.ErrProjectExists,
		project.ErrSocialPlatformAlreadyEnabled,
		project.ErrUserAlreadyInProject,
		user.ErrExistingUser,
		media.ErrFileAlreadyExists,
	):
		return &e.APIError{
			Status:  http.StatusConflict,
			Code:    e.ErrCodeConflict,
			Message: err.Error(),
		}

	// Status 422 Unprocessable Entity
	case e.MatchError(err,
		publisher.ErrPlatformSecretsNotSet,
		publisher.ErrUserSecretsNotSet,
		post.ErrPostNotDraft,
		post.ErrPostNotLinkedToAnyPlatform,
		publisher.ErrNoPublishersAssigned,
		post.ErrPostNotScheduled,
		post.ErrPostIsIdea,
		post.ErrPostIsNotIdea,
		post.ErrPostNotArchived,
		project.ErrSocialPlatformNotEnabled,
		project.ErrBasicRoleCannotBeRemoved,
		project.ErrUserNotInProject,
		project.ErrNoDefaultUserForPlatform,
	):
		return &e.APIError{
			Status:  http.StatusUnprocessableEntity,
			Code:    e.ErrCodeValidation,
			Message: err.Error(),
		}

	// Status 404 Gone
	case e.MatchError(err,
		post.ErrPostNotFound,
		post.ErrProjectNotFound,
		publisher.ErrSocialPlatformNotFound,
		project.ErrUserNotFound,
		user.ErrUserNotFound,
	):
		return &e.APIError{
			Status:  http.StatusGone,
			Code:    e.ErrCodeNotFound,
			Message: err.Error(),
		}

	// Default: Status 500 Internal Server Error
	default:
		return &e.APIError{
			Status:  http.StatusInternalServerError,
			Code:    e.ErrCodeInternal,
			Message: "Internal server error",
		}
	}
}
