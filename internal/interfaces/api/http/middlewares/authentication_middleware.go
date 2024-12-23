package middlewares

import (
	"context"
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/authentication"
	"github.com/pedrodcsjostrom/opencm/internal/utils/errors"
)

func AuthMiddleware(authenticator authentication.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID, err := r.Cookie("session_id")
			if err != nil || sessionID.Value == "" {
				errors.WriteHttpError(w, errors.NewUnauthorizedError("Not authenticated"))
				return
			}

			ctx := r.Context()
			fingerprint, ok := ctx.Value(DeviceFingerprintKey).(string)
			if !ok {
				errors.WriteHttpError(w, errors.NewUnauthorizedError("Device fingerprint required"))
				return
			}
			session, err := authenticator.Authenticate(ctx, sessionID.Value, fingerprint)
			if err != nil {
				errors.WriteHttpError(w, errors.NewUnauthorizedError("Invalid session"))
				return
			}
			ctx = context.WithValue(ctx, UserIDKey, session.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
