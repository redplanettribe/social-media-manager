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
	authMiddleware := middlewares.AuthMiddleware(authenticator)

	// Health check routes
	router.Handle("GET /health", ChainMiddlewares(http.HandlerFunc(healthCheckHandler.HealthCheck), middlewares.LoggingMiddleware))
	router.Handle("GET /health/auth", ChainMiddlewares(http.HandlerFunc(healthCheckHandler.HealthCheck), middlewares.LoggingMiddleware, authMiddleware))

	// User routes
	router.HandleFunc("POST /users", userHandler.SignUp)
	router.HandleFunc("POST /users/login", userHandler.Login)
	router.Handle("GET /users/{id}", ChainMiddlewares(http.HandlerFunc(userHandler.GetUser), middlewares.LoggingMiddleware, authMiddleware))
	router.Handle("GET /users/roles", ChainMiddlewares(http.HandlerFunc(userHandler.GetRoles), middlewares.LoggingMiddleware, authMiddleware))
	router.Handle("POST /users/roles", ChainMiddlewares(http.HandlerFunc(userHandler.AssignRoleToUser), middlewares.LoggingMiddleware, authMiddleware))

	return router
}

func ChainMiddlewares(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
