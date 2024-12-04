package middlewares

import (
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/authorization"
)

func AuthorizationMiddleware(authorizer authorization.Authorizer, requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(userIDKey)
			if userID == nil {
				http.Error(w, "no user id in context", http.StatusInternalServerError)
				return
			}

			err := authorizer.Authorize(r.Context(), userID.(string), requiredPermission)
			if err != nil {
				http.Error(w, "Forbidden: "+err.Error(), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
