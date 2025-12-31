// Package transaction_test executes integration tests for the transaction service.
// These tests require a real database connection to validate atomicity,
// idempotency, and concurrency control under realistic conditions.
package transaction

import (
	"context"
	"errors"
	"fmt"
	"ledger-project/internal/account"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Global variables for the test suite.
// Declaring them at the package level ensures they are accessible by all test functions.
var (
	svc     *Service
	accRepo *account.Repository
	pool    *pgxpool.Pool
)

// TestMain manages the lifecycle of the integration test suite.
// It handles environment loading, database connection setup, and teardown.
func TestMain(m *testing.M) {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// 1. Load environment variables from .env file
	// We use "../../.env" assuming the file is at the project root.
	if err := godotenv.Load("../../.env"); err != nil {
		fmt.Println("Warning: .env file not found, using system environment variables")
	}

	ctx := context.Background()

	// 2. Connect to the test database using the DATABASE_URL from .env
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		fmt.Println("Error: DATABASE_URL environment variable is not set")
		os.Exit(1)
	}

	// Increase max connections to prevent deadlocks during high-concurrency stress tests.
	dsn = dsn + "?pool_max_conns=30"

	testPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Printf("Failed to connect to test database: %v\n", err)
		os.Exit(1)
	}
	pool = testPool
	defer pool.Close()

	// 3. Initialize repositories and service with the real database pool
	accRepo = account.NewRepository(pool)
	txRepo := NewRepository(pool)
	svc = NewService(pool, accRepo, txRepo)

	// Run all tests in the package
	exitCode := m.Run()

	os.Exit(exitCode)
}

// TestExecuteTransfer_RaceCondition_Atomic verifies the system's resilience against
// concurrent race conditions. It ensures that when multiple requests attempt
// to debit the same account simultaneously, the database-level atomic
// constraints prevent the balance from ever becoming negative.
func TestExecuteTransfer_RaceCondition_Atomic(t *testing.T) {
	if pool == nil {
		t.Skip("Skipping integration test: database pool not initialized.")
	}

	// 1. SETUP: Prepare two accounts for the stress test.
	fromID := uuid.New()
	toID := uuid.New()

	// Start the source account with exactly $100.00
	setupAccount(t, fromID, 100)
	setupAccount(t, toID, 0)

	const (
		transferAmount = 20
		concurrency    = 10 // 10 simultaneous requests
	)

	// WaitGroup and Channel to synchronize goroutines and collect their results
	var wg sync.WaitGroup
	results := make(chan error, concurrency)

	// 2. EXECUTION: Launch 10 goroutines attempting to transfer $20 each simultaneously.
	// Only 5 should succeed (5 * 20 = 100), as the 6th will trigger insufficient funds.
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each request uses a unique Idempotency Key
			idKey := uuid.New()
			err := svc.ExecuteTransfer(context.Background(), fromID, toID, transferAmount, "Stress Test", idKey)
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	// 3. ANALYSIS: Count successes and expected business failures.
	successCount := 0
	insufficientFundsCount := 0

	for err := range results {
		if err == nil {
			successCount++
		} else if errors.Is(err, ErrInsufficientFunds) {
			insufficientFundsCount++
		}
	}

	// 4. ASSERTIONS: Verify the integrity of the ledger state.

	// We expect exactly 5 successes (Total $100 / $20 per transfer)
	if successCount != 5 {
		t.Errorf("Concurrency failure: expected 5 successes, but got %d", successCount)
	}

	// We expect exactly 5 failures with the specific ErrInsufficientFunds error
	if insufficientFundsCount != 5 {
		t.Errorf("Error handling failure: expected 5 insufficient funds errors, but got %d", insufficientFundsCount)
	}

	// Final verification: The account balance must be exactly zero.
	finalAcc, err := accRepo.GetByID(context.Background(), fromID)
	if err != nil {
		t.Fatalf("Failed to fetch final account state: %v", err)
	}

	if finalAcc.Balance != 0 {
		t.Errorf("Data integrity violation: final balance should be 0, but is %d", finalAcc.Balance)
	}
}

// setupAccount is a technical helper to ensure the database is in the correct state for a test case.
// It uses the global accRepo instance shared across the test package.
func setupAccount(t *testing.T, id uuid.UUID, balance int64) {
	t.Helper()

	// Using the real repository to insert test data.
	// This requires the Create method in internal/account/repository.go
	err := accRepo.Create(context.Background(), &account.Account{
		ID:      id,
		Balance: balance,
		Active:  true,
	})
	if err != nil {
		t.Fatalf("Failed to setup test account %s: %v", id, err)
	}
}
