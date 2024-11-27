package api

import (
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/handlers"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/auth"
)

// NewRouter creates and returns an http.Handler with all routes defined.
func NewRouter(
	healthCheckHandler *handlers.HealthHandler,
	authenticator auth.Authenticator,
) http.Handler {
	router := http.NewServeMux()
	// Define the middleware
	authMiddleware := middlewares.AuthMiddleware(authenticator)

	// Public routes
	router.HandleFunc("GET /health", healthCheckHandler.HealthCheck)

	// Private routes
	router.Handle("GET /health/auth", authMiddleware(http.HandlerFunc(healthCheckHandler.HealthCheck)))

	return router
}
