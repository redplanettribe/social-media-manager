package middlewares

import (
	"context"
	"log"
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/authentication"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(authenticator authentication.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID, err := r.Cookie("session_id")
			if err != nil {
				log.Printf("Error getting sessionID from cookie: %s", err)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ctx := r.Context()
			session, err := authenticator.Authenticate(ctx, sessionID.Value)
			if err != nil {
				log.Printf("Error authenticating: %s", err)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			ctx = context.WithValue(ctx, userIDKey, session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
