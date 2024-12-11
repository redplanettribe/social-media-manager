package api

import (
	"net/http"

	_ "github.com/pedrodcsjostrom/opencm/docs"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/handlers"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/authentication"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/authorization"
)

// NewRouter creates and returns an http.Handler with all routes defined.
func NewRouter(
	healthCheckHandler *handlers.HealthHandler,
	userHandler *handlers.UserHandler,
	projectHandler *handlers.ProjectHandler,
	authenticator authentication.Authenticator,
	authorizer authorization.Authorizer,
) http.Handler {
	router := http.NewServeMux()
	authenticationMiddleware := middlewares.AuthMiddleware(authenticator)

	router.Handle("/swagger/", httpSwagger.Handler(
        httpSwagger.URL("/swagger/doc.json"),
        httpSwagger.DeepLinking(true),
        httpSwagger.DocExpansion("none"),
        httpSwagger.DomID("swagger-ui"),
    ))

	// Apply CORS middleware to all routes
	router.Handle("/", middlewares.CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	// Health check routes
	router.Handle("GET /health", ChainMiddlewares(http.HandlerFunc(healthCheckHandler.HealthCheck),
		middlewares.CORSMiddleware,
		middlewares.LoggingMiddleware,
	))
	router.Handle("GET /health/auth", ChainMiddlewares(http.HandlerFunc(healthCheckHandler.HealthCheck),
		authenticationMiddleware,
		middlewares.CORSMiddleware,
		middlewares.LoggingMiddleware,
	))

	// User routes
	router.Handle("POST /users", ChainMiddlewares(http.HandlerFunc(userHandler.SignUp),
		middlewares.CORSMiddleware,
		middlewares.AddDeviceFingerprint,
		middlewares.LoggingMiddleware,
	))
	router.Handle("POST /users/login", ChainMiddlewares(http.HandlerFunc(userHandler.Login),
		middlewares.CORSMiddleware,
		middlewares.AddDeviceFingerprint,
		middlewares.LoggingMiddleware,
	))
	router.Handle("POST /users/logout", ChainMiddlewares(http.HandlerFunc(userHandler.Logout),
	middlewares.CORSMiddleware,
	middlewares.AddDeviceFingerprint,
	middlewares.LoggingMiddleware,
))
	router.Handle("GET /users/me", ChainMiddlewares(http.HandlerFunc(userHandler.GetUser),
		middlewares.CORSMiddleware,
		middlewares.AuthorizationMiddleware(authorizer, "read:users"),
		authenticationMiddleware,
		middlewares.AddDeviceFingerprint,
		middlewares.LoggingMiddleware,
	))
	router.Handle("GET /users/roles", ChainMiddlewares(http.HandlerFunc(userHandler.GetRoles),
		middlewares.CORSMiddleware,
		middlewares.AuthorizationMiddleware(authorizer, "read:roles"),
		authenticationMiddleware,
		middlewares.AddDeviceFingerprint,
		middlewares.LoggingMiddleware,
	))
	router.Handle("POST /users/roles", ChainMiddlewares(http.HandlerFunc(userHandler.AssignRoleToUser),
		middlewares.CORSMiddleware,
		middlewares.AuthorizationMiddleware(authorizer, "write:roles"),
		authenticationMiddleware,
		middlewares.AddDeviceFingerprint,
		middlewares.LoggingMiddleware,
	))
	router.Handle("DELETE /users/roles", ChainMiddlewares(http.HandlerFunc(userHandler.RemoveRoleFromUser),
		middlewares.CORSMiddleware,
		middlewares.AuthorizationMiddleware(authorizer, "delete:roles"),
		authenticationMiddleware,
		middlewares.AddDeviceFingerprint,
		middlewares.LoggingMiddleware,
	))

	// Project routes
	router.Handle("POST /projects", ChainMiddlewares(http.HandlerFunc(projectHandler.CreateProject),
		middlewares.CORSMiddleware,
		middlewares.AuthorizationMiddleware(authorizer, "write:projects"),
		authenticationMiddleware,
		middlewares.AddDeviceFingerprint,
		middlewares.LoggingMiddleware,
	))

	return router
}

func ChainMiddlewares(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
