// Package main is the entry point for the Ledger Service API.
// It handles infrastructure initialization, dependency injection, and server lifecycle.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ledger-project/internal/account"
	"ledger-project/internal/api"
	"ledger-project/internal/database"
	"ledger-project/internal/transaction"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// 1. Logger Initialization
	// We use ConsoleWriter for human-readable logs during development.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	log.Info().Msg("Initializing Ledger Service...")

	// 2. Environment Variables
	// Loads sensitive data like DATABASE_URL from a separate .env file.
	// This follows the security requirement of keeping credentials out of the codebase.
	if err := godotenv.Load("../../.env"); err != nil {
		log.Warn().Msg(".env file not found, falling back to system environment variables")
	}

	// 3. Infrastructure: Database Connection
	// Using pgxpool for efficient connection management.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbPool := database.ConnectDB()
	defer func() {
		log.Info().Msg("Closing database connection pool...")
		dbPool.Close()
	}()

	// 4. Infrastructure: Database Migrations & Seeding
	// Ensures the schema is up-to-date and populated with initial dev data.
	if err := database.RunMigrations(ctx, dbPool); err != nil {
		log.Fatal().Err(err).Msg("Critical Failure: Database migrations failed")
	}

	if err := SeedInitialData(ctx, dbPool); err != nil {
		log.Error().Err(err).Msg("Seed data warning (check if data already exists)")
	}

	// 5. Domain Layer: Repositories & Services
	// Applying Dependency Injection to allow easier testing and loose coupling.
	accRepo := account.NewRepository(dbPool)
	txRepo := transaction.NewRepository(dbPool)

	// The transaction service manages the atomic business logic of fund transfers.
	txService := transaction.NewService(dbPool, accRepo, txRepo)

	// 6. Interface Layer: API Handlers and Router
	handler := api.NewHandler(txService)
	router := api.NewRouter(handler)

	// 7. HTTP Server Lifecycle Management
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Server execution in a separate goroutine to allow for graceful shutdown listening.
	go func() {
		log.Info().Str("port", port).Msg("Ledger API is online and listening")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed to start")
		}
	}()

	// 8. Graceful Shutdown
	// Blocks until an interrupt signal is received (CTRL+C or SIGTERM).
	<-ctx.Done()
	log.Info().Msg("Shutting down gracefully...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Ledger Service stopped")
}
