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
	userHandler *handlers.UserHandler,
	authenticator auth.Authenticator,
) http.Handler {
	router := http.NewServeMux()
	authMiddleware := middlewares.AuthMiddleware(authenticator) // make different stacks of middlewares

	// Health check routes
	router.HandleFunc("GET /health", healthCheckHandler.HealthCheck)
	router.Handle("GET /health/auth", authMiddleware(http.HandlerFunc(healthCheckHandler.HealthCheck)))

	// User routes
	router.HandleFunc("POST /users", userHandler.SignUp)
	router.HandleFunc("POST /users/login", userHandler.Login)
	router.Handle("GET /users/{id}", authMiddleware(http.HandlerFunc(userHandler.GetUser)))

	return router
}
