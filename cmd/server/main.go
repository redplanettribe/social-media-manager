package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/pedrodcsjostrom/opencm/internal/domain/project"
	"github.com/pedrodcsjostrom/opencm/internal/domain/user"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/encrypting"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/persistence/postgres"
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

	dbConn, err := pgx.Connect(ctx, dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := dbConn.Close(ctx); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	authenticator := authentication.NewAuthenticator(session.NewManager(postgres.NewSessionRepository(dbConn)))

	healthHandler := handlers.NewHealthHandler()

	userRepo := postgres.NewUserRepository(dbConn)
	passworHasher := encrypting.NewHasher()
	sessionRepo := postgres.NewSessionRepository(dbConn)
	sessionManager := session.NewManager(sessionRepo)
	userService := user.NewService(userRepo, sessionManager, passworHasher)
	userHandler := handlers.NewUserHandler(userService)

	projctRepo := postgres.NewProjectRepository(dbConn)
	projectService := project.NewService(projctRepo)
	projectHandler := handlers.NewProjectHandler(projectService)

	authorizer := authorization.NewAuthorizer(authorization.GetPermissions(), userService.GetUserAppRoles)
	httpRouter := api.NewRouter(healthHandler, userHandler, projectHandler, authenticator, authorizer)

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // temporary configuration test
	}

	server := &http.Server{
		Addr:      ":" + cfg.App.Port,
		Handler:   httpRouter,
		TLSConfig: tlsConfig,
	}

	log.Printf("Server is running on port %s", cfg.App.Port)
	if err := server.ListenAndServeTLS(cfg.SSL.CertPath, cfg.SSL.KeyPath); err != nil {
		log.Fatal(err)
	}
}
