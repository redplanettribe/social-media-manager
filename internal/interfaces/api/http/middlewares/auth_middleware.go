package middlewares

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/auth"
)

type contextKey string

const UserIDKey contextKey = "userID"
const TeamIDKey contextKey = "teamID"

type Middleware func(http.HandlerFunc) http.HandlerFunc

func AuthMiddleware(authenticator auth.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			id, err := authenticator.Authenticate(r.Header.Get("Authorization"))
			if err != nil {
				log.Printf("Error authenticating: %s", err)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
			log.Println(r.Method, r.URL.Path, time.Since(start))
		})
	}
}
