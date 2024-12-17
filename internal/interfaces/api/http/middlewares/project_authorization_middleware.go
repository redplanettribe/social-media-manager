package middlewares

import (
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/authorization"
)

func ProjectAuthorizationMiddleware(authorizer authorization.ProjectAuthorizer, requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID := ctx.Value(UserIDKey)
			if userID == nil {
				http.Error(w, "no user id in context", http.StatusInternalServerError)
				return
			}
			projectID := r.PathValue("project_id")
			if projectID == "" {
				http.Error(w, "project id not found in path value", http.StatusBadRequest)
				return
			}

			err := authorizer.Authorize(r.Context(), userID.(string), projectID, requiredPermission)
			if err != nil {
				http.Error(w, "Forbidden: "+err.Error(), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
