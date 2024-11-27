package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load configuration: %v", err)
	}

	// Initialize database connection
	dbConnStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)

	dbConn, err := pgx.Connect(context.Background(), dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close(context.Background())

	// logger := logging.NewLogger(&cfg.Logger)

	// // Initialize repositories
	// postRepo := postgres.NewPostRepository(db)

	// // Initialize domain services
	// postService := post.NewService(postRepo)

	// // Initialize application layer handlers
	// createPostHandler := commands.NewCreatePostHandler(postService)

	// // Simple command bus implementation
	// commandBus := &SimpleCommandBus{
	// 	handlers: map[string]interface{}{
	// 		"CreatePostCommand": createPostHandler,
	// 	},
	// }

	// // Initialize authenticator
	// authenticator := auth.NewJWTAuthenticator(cfg.JWT.SecretKey)

	// // Initialize HTTP handlers
	// postHandler := handlers.NewPostHandler(commandBus)

	// Set up router
	// httpRouter := router.NewRouter(postHandler, authenticator)

	// Start the server
	// log.Printf("Server is running on port %s", cfg.App.Port)
	// if err := http.ListenAndServe(":"+cfg.App.Port, httpRouter); err != nil {
	// 	log.Fatal(err)
	// }
}
