package postgres_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/redplanettribe/social-media-manager/internal/infrastructure/config"
)

var dbPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()
	// 1. Setup database connection (e.g. local or Docker)
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

	pool, err := pgxpool.NewWithConfig(ctx, dbConf)
	if err != nil {
		log.Fatal(err)
	}
	dbPool = pool

	// 2. Optionally run migrations
	// migrateDB(connStr)

	// 3. Run all tests
	code := m.Run()

	// 4. Close pool, clean up
	pool.Close()

	// 5. Exit with test status code
	os.Exit(code)
}

// WAY TO TEST THE REPOS WITH THE DATABASE

// func TestPostRepository_FindScheduledReadyPosts(t *testing.T) {
// 	repo := postgres.NewPostRepository(dbPool)

// 	// Find the post
// 	posts, err := repo.FindScheduledReadyPosts(context.Background(), 0, 10)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, posts)
// }
