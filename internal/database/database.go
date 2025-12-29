// Package database handles the lifecycle of the database engine connection.
package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

// ConnectDB initializes a PostgreSQL connection using environment variables.
// It loads the .env file, parses the DB_URL, and validates the connection with a Ping.
func ConnectDB() *pgx.Conn {
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
	config, err := pgx.ParseConfig(connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Connection error: %v\n", err)
		os.Exit(1)
	}

	if err = conn.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Database unreachable: %v\n", err)
		os.Exit(1)
	}

	return conn
}
