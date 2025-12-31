// Package database handles the lifecycle of the database engine connection.
package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool" // Import the pool sub-package
	"github.com/joho/godotenv"
)

// ConnectDB initializes a PostgreSQL connection pool using environment variables.
// It loads the .env file and establishes a resilient pool for high-concurrency ledger operations.
func ConnectDB() *pgxpool.Pool {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found, using system env")
	}

	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		fmt.Fprintf(os.Stderr, "Critical: DB_URL environment variable is not set\n")
		os.Exit(1)
	}

	ctx := context.Background()

	// NewPool creates a connection pool, which is much better for production than a single connection.
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	// Validate the connection
	if err = pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Database unreachable: %v\n", err)
		os.Exit(1)
	}

	return pool
}