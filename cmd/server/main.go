package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/pedrodcsjostrom/opencm/docs"
	"github.com/pedrodcsjostrom/opencm/internal/domain/platform"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
	"github.com/pedrodcsjostrom/opencm/internal/domain/project"
	"github.com/pedrodcsjostrom/opencm/internal/domain/user"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/encrypting"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/persistence/postgres"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/scheduler"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/server"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/session"
	api "github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/handlers"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/authentication"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/authorization"
)

// @title OpenCM API
// @version 1.0
// @description This is the API server for OpenCM
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https
func main() {
	ctx := context.Background()
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load configuration: %v", err)
	}

	// // Initialize database connection
	dbConnStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)
	dbConf, err := pgxpool.ParseConfig(dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	dbConf.MaxConns = 25                     // Maximum number of connections in the pool
	dbConf.MinConns = 5                      // Minimum number of connections to keep open
	dbConf.MaxConnLifetime = 5 * time.Minute // Maximum lifetime of a connection
	dbConf.MaxConnIdleTime = 1 * time.Minute // Maximum idle time of a connection

	dbPool, err := pgxpool.NewWithConfig(ctx, dbConf)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	authenticator := authentication.NewAuthenticator(session.NewManager(postgres.NewSessionRepository(dbPool)))

	healthHandler := handlers.NewHealthHandler()

	userRepo := postgres.NewUserRepository(dbPool)
	passworHasher := encrypting.NewHasher()
	sessionRepo := postgres.NewSessionRepository(dbPool)
	sessionManager := session.NewManager(sessionRepo)
	userService := user.NewService(userRepo, sessionManager, passworHasher)
	userHandler := handlers.NewUserHandler(userService)

	projectRepo := postgres.NewProjectRepository(dbPool)
	projectService := project.NewService(projectRepo, userRepo)
	projectHandler := handlers.NewProjectHandler(projectService)

	postRepo := postgres.NewPostRepository(dbPool)
	postService := post.NewService(postRepo)
	postHandler := handlers.NewPostHandler(postService)

	platformRepo := postgres.NewPlatformRepository(dbPool)
	platformService := platform.NewService(platformRepo)
	platformHandler := handlers.NewPlatformHandler(platformService)

	appAuthorizer := authorization.NewAppAuthorizer(authorization.GetAppPermissions(), userService.GetUserAppRoles)
	projectAuthorizer := authorization.NewTeamAthorizer(authorization.GetTeamPermissions(), projectService.GetUserRoles)
	httpRouter := api.NewRouter(
		healthHandler,
		userHandler,
		projectHandler,
		postHandler,
		platformHandler,
		authenticator,
		appAuthorizer,
		projectAuthorizer,
	)

	// Start the post scheduler
	scheduler := scheduler.NewPostScheduler(postService, projectService, &cfg.Scheduler)
	scheduler.Start(ctx)

	// Start the Server
	server := server.NewHttpServer(cfg, httpRouter)
	server.Serve()
}
