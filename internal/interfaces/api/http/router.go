package api

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/handlers"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/middlewares"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/authentication"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/authorization"
)

type middlewareStack []func(http.Handler) http.Handler

func (s middlewareStack) Chain(h http.Handler) http.Handler {
	return ChainMiddlewares(h, s...)
}

type Router struct {
	*http.ServeMux
	baseStack     middlewareStack
	authStack     middlewareStack
	authenticator authentication.Authenticator
	authorizer    authorization.Authorizer
}

func NewRouter(
	healthCheckHandler *handlers.HealthHandler,
	userHandler *handlers.UserHandler,
	projectHandler *handlers.ProjectHandler,
	authenticator authentication.Authenticator,
	authorizer authorization.Authorizer,
) http.Handler {
	r := &Router{
		ServeMux:      http.NewServeMux(),
		authenticator: authenticator,
		authorizer:    authorizer,
	}

	// Middleware stacks
	r.baseStack = middlewareStack{
		middlewares.LoggingMiddleware,
		middlewares.CORSMiddleware,
		middlewares.AddDeviceFingerprint,
	}

	authMiddleware := middlewares.AuthMiddleware(authenticator)
	r.authStack = append(r.baseStack, authMiddleware)

	// Setup routes
	r.setupSwagger()
	r.setupHealthRoutes(healthCheckHandler)
	r.setupUserRoutes(userHandler)
	r.setupProjectRoutes(projectHandler)

	return r
}

func (r *Router) setupSwagger() {
	r.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))
}

func (r *Router) setupHealthRoutes(h *handlers.HealthHandler) {
	r.Handle("GET /health", r.baseStack.Chain(
		http.HandlerFunc(h.HealthCheck),
	))
	r.Handle("GET /health/auth", r.authStack.Chain(
		http.HandlerFunc(h.HealthCheck),
	))
}

func (r *Router) setupUserRoutes(h *handlers.UserHandler) {
	r.Handle("POST /users", r.baseStack.Chain(
		http.HandlerFunc(h.SignUp),
	))
	r.Handle("POST /users/login", r.baseStack.Chain(
		http.HandlerFunc(h.Login),
	))
	r.Handle("POST /users/logout", r.baseStack.Chain(
		http.HandlerFunc(h.Logout),
	))

	// Protected routes
	r.Handle("GET /users/me", r.permissions("read:users").Chain(
		http.HandlerFunc(h.GetUser),
	))
	r.Handle("GET /users/roles", r.permissions("read:roles").Chain(
		http.HandlerFunc(h.GetRoles),
	))
	r.Handle("POST /users/roles", r.permissions("write:roles").Chain(
		http.HandlerFunc(h.AssignRoleToUser),
	))
	r.Handle("DELETE /users/roles", r.permissions("delete:roles").Chain(
		http.HandlerFunc(h.RemoveRoleFromUser),
	))
}

func (r *Router) setupProjectRoutes(h *handlers.ProjectHandler) {
	r.Handle("POST /projects", r.permissions("write:projects").Chain(
		http.HandlerFunc(h.CreateProject),
	))
	r.Handle("GET /projects", r.permissions("read:projects").Chain(
		http.HandlerFunc(h.ListProjects),
	))
}

// permissions returns a middleware stack that checks if the user has the required permission for the desired action
func (r *Router) permissions(permission string) middlewareStack {
	return append(r.authStack,
		middlewares.AuthorizationMiddleware(r.authorizer, permission),
	)
}

func ChainMiddlewares(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
