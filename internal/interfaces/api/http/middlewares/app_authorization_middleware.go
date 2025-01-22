package middlewares

import (
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/authorization"
	"github.com/pedrodcsjostrom/opencm/internal/utils/errors"
)

func AppAuthorizationMiddleware(authorizer authorization.AppAuthorizer, requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(UserIDKey)
			if userID == nil {
				errors.WriteHttpError(w, errors.NewInternalError("no user id in context"))
				return
			}

			err := authorizer.Authorize(r.Context(), userID.(string), requiredPermission)
			if err != nil {
				errors.WriteHttpError(w, errors.NewForbiddenError("Not authorized to perform this action"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
