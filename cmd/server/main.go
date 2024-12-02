package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/pedrodcsjostrom/opencm/internal/domain/user"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/encrypting"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/persistence/postgres"
	api "github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/handlers"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/auth"
)

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
	defer dbConn.Close(ctx)

	authenticator := auth.NewJWTAuthenticator(cfg.JWT.SecretKey)

	healthHandler := handlers.NewHealthHandler()

	userRepo := postgres.NewUserRepository(dbConn)
	passworHasher := encrypting.NewHasher()
	userService := user.NewService(userRepo, passworHasher)
	userHandler := handlers.NewUserHandler(userService)

	// Set up router
	httpRouter := api.NewRouter(healthHandler, userHandler, authenticator)

	// Start the server
	log.Printf("Server is running on port %s", cfg.App.Port)
	if err := http.ListenAndServe(":"+cfg.App.Port, httpRouter); err != nil {
		log.Fatal(err)
	}
}
