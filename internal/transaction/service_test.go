// Package transaction contains unit tests for the transaction service.
package transaction

import (
	"context"
	"errors"
	"ledger-project/internal/account"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// --- MOCKS ---

type MockAccountRepo struct {
	Balance int64
}

func (m *MockAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	return &account.Account{ID: id, Balance: m.Balance}, nil
}

func (m *MockAccountRepo) UpdateBalance(ctx context.Context, tx pgx.Tx, id uuid.UUID, amount int64) error {
	return nil
}

type MockTxRepo struct{}

func (m *MockTxRepo) CreateIdempotencyKey(ctx context.Context, tx pgx.Tx, key uuid.UUID) error {
	return nil
}

func (m *MockTxRepo) Create(ctx context.Context, tx pgx.Tx, t *Transaction) error {
	return nil
}

// MockTx simulates a database transaction (pgx.Tx)
type MockTx struct {
	pgx.Tx // Embed to satisfy the interface
}

func (m *MockTx) Commit(ctx context.Context) error   { return nil }
func (m *MockTx) Rollback(ctx context.Context) error { return nil }

// MockPool simulates the connection pool
type MockPool struct{}

func (m *MockPool) Begin(ctx context.Context) (pgx.Tx, error) {
	return &MockTx{}, nil // Returns our fake transaction
}

// --- TESTS ---

func TestExecuteTransfer_InsufficientFunds(t *testing.T) {
	// 1. Setup with the new MockPool instead of nil
	mockAcc := &MockAccountRepo{Balance: 50}
	mockTx := &MockTxRepo{}
	mockPool := &MockPool{}

	// Now svc.pool.Begin(ctx) will call MockPool.Begin and NOT panic!
	svc := NewService(mockPool, mockAcc, mockTx)

	fromID := uuid.New()
	toID := uuid.New()
	idKey := uuid.New()

	// 2. Execution
	err := svc.ExecuteTransfer(context.Background(), fromID, toID, 100, "Test", idKey)

	// 3. Assertion
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Errorf("Expected ErrInsufficientFunds, got: %v", err)
	}
}
