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
	baseStack         middlewareStack
	authStack         middlewareStack
	authenticator     authentication.Authenticator
	appAuthorizer     authorization.AppAuthorizer
	projectAuthorizer authorization.ProjectAuthorizer
}

func NewRouter(
	healthCheckHandler *handlers.HealthHandler,
	userHandler *handlers.UserHandler,
	projectHandler *handlers.ProjectHandler,
	postHandler *handlers.PostHandler,
	platformHandler *handlers.PlatformHandler,
	mediaHandler *handlers.MediaHandler,
	authenticator authentication.Authenticator,
	appAuthorizer authorization.AppAuthorizer,
	projectAuthorizer authorization.ProjectAuthorizer,
) http.Handler {
	r := &Router{
		ServeMux:          http.NewServeMux(),
		authenticator:     authenticator,
		appAuthorizer:     appAuthorizer,
		projectAuthorizer: projectAuthorizer,
	}

	// Middleware stacks
	r.baseStack = middlewareStack{
		middlewares.LoggingMiddleware,
		middlewares.AddDeviceFingerprint,
		middlewares.CORSMiddleware,
	}

	authMiddleware := middlewares.AuthMiddleware(authenticator)
	r.authStack = append(r.baseStack, authMiddleware)

	// Setup routes
	r.setUpCors()
	r.setupSwagger()
	r.setupHealthRoutes(healthCheckHandler)
	r.setupUserRoutes(userHandler)
	r.setupProjectRoutes(projectHandler)
	r.setupPostRoutes(postHandler)
	r.setupPublisherRoutes(platformHandler)
	r.setupMediaRoutes(mediaHandler)

	return r
}

func (r *Router) setUpCors() {
	r.Handle("/", middlewares.CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

}

func (r *Router) setupSwagger() {
	r.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))
}

/*HEALTH ROUTES*/
func (r *Router) setupHealthRoutes(h *handlers.HealthHandler) {
	r.Handle("GET /health", r.baseStack.Chain(
		http.HandlerFunc(h.HealthCheck),
	))
	r.Handle("GET /health/auth", r.authStack.Chain(
		http.HandlerFunc(h.HealthCheck),
	))
}

/*USER ROUTES*/
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
	r.Handle("GET /users/me", r.appPermissions("read:users").Chain(
		http.HandlerFunc(h.GetUser),
	))
	r.Handle("GET /users/roles", r.appPermissions("read:roles").Chain(
		http.HandlerFunc(h.GetRoles),
	))
	r.Handle("POST /users/roles", r.appPermissions("write:roles").Chain(
		http.HandlerFunc(h.AssignRoleToUser),
	))
	r.Handle("DELETE /users/roles", r.appPermissions("delete:roles").Chain(
		http.HandlerFunc(h.RemoveRoleFromUser),
	))
}

/*PROJECT ROUTES*/
func (r *Router) setupProjectRoutes(h *handlers.ProjectHandler) {
	r.Handle("POST /projects", r.appPermissions("write:projects").Chain(
		http.HandlerFunc(h.CreateProject),
	))
	r.Handle("POST /projects/{project_id}/add", r.projectPermissions("write:projects").Chain(
		http.HandlerFunc(h.AddUserToProject),
	))
	r.Handle("POST /projects/{project_id}/enable-social-platform/{platform_id}", r.projectPermissions("write:projects").Chain(
		http.HandlerFunc(h.EnableSocialPlatform),
	))
	r.Handle("PATCH /projects/{project_id}/time-zone", r.projectPermissions("write:projects").Chain(
		http.HandlerFunc(h.SetTimeZone),
	))
	r.Handle("PATCH /projects/{project_id}/time-slots", r.projectPermissions("write:projects").Chain(
		http.HandlerFunc(h.AddTimeSlot),
	))
	r.Handle("PATCH /projects/{project_id}/default-user/{user_id}", r.projectPermissions("write:projects").Chain(
		http.HandlerFunc(h.SetDefaultUser),
	))
	r.Handle("GET /projects", r.appPermissions("read:projects").Chain(
		http.HandlerFunc(h.ListProjects),
	))
	r.Handle("GET /projects/{project_id}/social-platforms", r.projectPermissions("read:projects").Chain(
		http.HandlerFunc(h.GetEnabledSocialPlatforms),
	))
	r.Handle("GET /projects/{project_id}", r.projectPermissions("read:projects").Chain(
		http.HandlerFunc(h.GetProject),
	))
}

/*POST ROUTES*/
func (r *Router) setupPostRoutes(h *handlers.PostHandler) {
	r.Handle("POST /posts/{project_id}/add", r.projectPermissions("write:posts").Chain(
		http.HandlerFunc(h.CreatePost),
	))
	r.Handle("POST /posts/{project_id}/{post_id}/platforms/{platform_id}", r.projectPermissions("write:posts").Chain(
		http.HandlerFunc(h.AddSocialMediaPublisherPlatform),
	))
	r.Handle("PATCH /posts/{project_id}/{post_id}/schedule", r.projectPermissions("write:posts").Chain(
		http.HandlerFunc(h.SchedulePost),
	))
	r.Handle("PATCH /posts/{project_id}/{post_id}/archive", r.projectPermissions("write:posts").Chain(
		http.HandlerFunc(h.ArchivePost),
	))
	r.Handle("PATCH /posts/{project_id}/{post_id}/enqueue", r.projectPermissions("write:posts").Chain(
		http.HandlerFunc(h.AddPostToProjectQueue),
	))
	r.Handle("PATCH /posts/{project_id}/queue/move", r.projectPermissions("write:posts").Chain(
		http.HandlerFunc(h.MovePostInQueue),
	))
	r.Handle("GET /posts/{project_id}/{post_id}", r.projectPermissions("read:posts").Chain(
		http.HandlerFunc(h.GetPost),
	))
	r.Handle("GET /posts/{project_id}", r.projectPermissions("read:posts").Chain(
		http.HandlerFunc(h.ListProjectPosts),
	))
	r.Handle("GET /posts/{project_id}/queue", r.projectPermissions("read:posts").Chain(
		http.HandlerFunc(h.GetProjectQueuedPosts),
	))
	r.Handle("GET /posts", r.appPermissions("read:posts").Chain(
		http.HandlerFunc(h.GetAvailablePostTypes),
	))
	r.Handle("DELETE /posts/{project_id}/{post_id}", r.projectPermissions("delete:posts").Chain(
		http.HandlerFunc(h.DeletePost),
	))
}

/*PUBLISHER ROUTES*/
func (r *Router) setupPublisherRoutes(h *handlers.PlatformHandler) {
	r.Handle("POST /publishers/{project_id}/{post_id}/{social_network_id}", r.projectPermissions("write:publishers").Chain(
		http.HandlerFunc(h.PublishPostToSocialNetwork),
	))
	r.Handle("POST /publishers/{project_id}/{post_id}", r.projectPermissions("write:publishers").Chain(
		http.HandlerFunc(h.PublishPostToAssignedSocialNetworks),
	))
	r.Handle("GET /publishers", r.appPermissions("read:publishers").Chain(
		http.HandlerFunc(h.GetAvailableSocialNetworks),
	))
	r.Handle("PATCH /publishers/{project_id}/platform-secrets", r.projectPermissions("write:publishers").Chain(
		http.HandlerFunc(h.AddPlatformSecret),
	))
	r.Handle("PATCH /publishers/{project_id}/user-secrets", r.projectPermissions("write:publishers").Chain(
		http.HandlerFunc(h.AddUserPlatformSecret),
	))
}

/*MEDIA ROUTES*/
func (r *Router) setupMediaRoutes(h *handlers.MediaHandler) {
	r.Handle("POST /media/{project_id}/{post_id}", r.projectPermissions("write:media").Chain(
		http.HandlerFunc(h.UploadMedia),
	))
	r.Handle("POST /media/{project_id}/link-to-post", r.projectPermissions("write:media").Chain(
		http.HandlerFunc(h.LinkMediaToPublishPost),
	))
	r.Handle("GET /media/{project_id}/{post_id}/{file_name}", r.projectPermissions("read:media").Chain(
		http.HandlerFunc(h.GetMediaFile),
	))
}

// appPermissions returns a middleware stack that checks if the user has the required permission for the desired action
func (r *Router) appPermissions(permission string) middlewareStack {
	return append(r.authStack,
		middlewares.AppAuthorizationMiddleware(r.appAuthorizer, permission),
	)
}

func (r *Router) projectPermissions(permission string) middlewareStack {
	return append(r.authStack,
		middlewares.ProjectAuthorizationMiddleware(r.projectAuthorizer, permission),
	)
}

func ChainMiddlewares(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
