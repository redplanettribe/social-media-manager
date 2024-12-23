package middlewares

import (
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/authorization"
	e "github.com/pedrodcsjostrom/opencm/internal/utils/errors"
)

func ProjectAuthorizationMiddleware(authorizer authorization.ProjectAuthorizer, requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID := ctx.Value(UserIDKey)
			if userID == nil {
				e.WriteHttpError(w, e.NewInternalError("no user id in context"))
				return
			}
			projectID := r.PathValue("project_id")
			if projectID == "" {
				e.WriteHttpError(w, e.NewInternalError("no project id in path"))
				return
			}

			err := authorizer.Authorize(r.Context(), userID.(string), projectID, requiredPermission)
			if err != nil {
				e.WriteHttpError(w, e.NewUnauthorizedError("Not authorized to perform this action"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
