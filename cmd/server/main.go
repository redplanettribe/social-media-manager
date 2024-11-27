package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/application/commands"
	"github.com/pedrodcsjostrom/opencm/internal/application/interfaces"
	"github.com/pedrodcsjostrom/opencm/internal/infrastructure/config"
	api "github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/api/http/handlers"
	"github.com/pedrodcsjostrom/opencm/internal/interfaces/auth"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load configuration: %v", err)
	}

	// // Initialize database connection
	// dbConnStr := fmt.Sprintf(
	// 	"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
	// 	cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	// )

	// dbConn, err := pgx.Connect(context.Background(), dbConnStr)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer dbConn.Close(context.Background())

	// Initialize authenticator
	authenticator := auth.NewJWTAuthenticator(cfg.JWT.SecretKey)

	// Initialize health check handler
	healthHandler := handlers.NewHealthHandler()

	// Set up router
	httpRouter := api.NewRouter(healthHandler, authenticator)

	// Start the server
	log.Printf("Server is running on port %s", cfg.App.Port)
	if err := http.ListenAndServe(":"+cfg.App.Port, httpRouter); err != nil {
		log.Fatal(err)
	}
}

type SimpleCommandBus struct {
	handlers map[string]interface{}
}

func (bus *SimpleCommandBus) Dispatch(ctx context.Context, cmd interfaces.Command) error {
	switch c := cmd.(type) {
	case *commands.CreatePostCommand:
		handler := bus.handlers["CreatePostCommand"].(*commands.CreatePostHandler)
		return handler.Handle(ctx, c)
	// Handle other commands...
	default:
		return errors.New("unknown command")
	}
}
