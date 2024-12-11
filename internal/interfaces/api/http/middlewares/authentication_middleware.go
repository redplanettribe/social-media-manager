package middlewares

import (
	"context"
	"log"
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/authentication"
)

func AuthMiddleware(authenticator authentication.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID, err := r.Cookie("session_id")
			if err != nil || sessionID.Value == "" {
				log.Printf("Error getting sessionID from cookie: %s", err)
				http.Error(w, "Session required", http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			fingerprint, ok := ctx.Value(DeviceFingerprintKey).(string)
			if !ok {
				log.Printf("Error getting fingerprint from context")
				http.Error(w, "Fingerprint required", http.StatusUnauthorized)
				return
			}
			session, err := authenticator.Authenticate(ctx, sessionID.Value, fingerprint)
			if err != nil {
				log.Printf("Error authenticating: %s", err)
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
			ctx = context.WithValue(ctx, UserIDKey, session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
