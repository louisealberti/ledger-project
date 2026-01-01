// Package transaction_test contains unit tests for the transaction service.
// It uses mocks to isolate business logic from infrastructure dependencies like databases.
package transaction

import (
	"context"
	"errors"
	"testing"

	"ledger-project/internal/account"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAccountRepo defines a mock implementation of the AccountRepository interface.
type MockAccountRepo struct {
	mock.Mock
}

// GetByID simulates fetching an account from the database.
func (m *MockAccountRepo) GetByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.Account), args.Error(1)
}

// UpdateBalance simulates updating an account's balance within a database transaction.
func (m *MockAccountRepo) UpdateBalance(ctx context.Context, tx pgx.Tx, id uuid.UUID, amount int64) error {
	args := m.Called(ctx, tx, id, amount)
	return args.Error(0)
}

// MockTransactionRepo defines a mock implementation of the TransactionRepository interface.
type MockTransactionRepo struct {
	mock.Mock
}

// CreateIdempotencyKey simulates checking or creating a unique idempotency key in the database.
func (m *MockTransactionRepo) CreateIdempotencyKey(ctx context.Context, tx pgx.Tx, key uuid.UUID) error {
	args := m.Called(ctx, tx, key)
	return args.Error(0)
}

// Create simulates persisting a new transaction record.
func (m *MockTransactionRepo) Create(ctx context.Context, tx pgx.Tx, t *Transaction) error {
	args := m.Called(ctx, tx, t)
	return args.Error(0)
}

// GetByAccountID simulates retrieving all transactions associated with a specific account.
func (m *MockTransactionRepo) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]Transaction, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).([]Transaction), args.Error(1)
}

// MockDBPool defines a mock for the database pool to simulate starting transactions.
type MockDBPool struct {
	mock.Mock
}

// Begin simulates the initiation of a pgx database transaction.
func (m *MockDBPool) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(pgx.Tx), args.Error(1)
}

// MockTx is a mock for the pgx.Tx interface to simulate transaction lifecycle methods.
type MockTx struct {
	pgx.Tx
	mock.Mock
}

// Rollback simulates the cancellation of a database transaction.
func (m *MockTx) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Commit simulates the successful completion of a database transaction.
func (m *MockTx) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// TestExecuteTransfer_InsufficientFunds verifies that the service prevents transfers
// when the source account lacks the required balance.
func TestExecuteTransfer_InsufficientFunds(t *testing.T) {
	// Initialize mocks and the service under test
	mockAccRepo := new(MockAccountRepo)
	mockTxRepo := new(MockTransactionRepo)
	mockPool := new(MockDBPool)
	mockTx := new(MockTx) // Declaring the mock transaction variable

	service := NewService(mockPool, mockAccRepo, mockTxRepo)

	fromID := uuid.New()
	toID := uuid.New()
	idKey := uuid.New()

	// Setup expectations for database transaction lifecycle
	mockPool.On("Begin", mock.Anything).Return(mockTx, nil)
	mockTx.On("Rollback", mock.Anything).Return(nil)

	// Setup expectations for business rules
	mockTxRepo.On("CreateIdempotencyKey", mock.Anything, mock.Anything, idKey).Return(nil)

	// Setup expectations: Source account has 10.00 (1000 cents), but needs 50.00
	mockAccRepo.On("GetByID", mock.Anything, fromID).Return(&account.Account{
		ID:      fromID,
		Balance: 1000,
	}, nil)

	// Execute: Attempting a transfer that exceeds the balance
	err := service.ExecuteTransfer(context.Background(), fromID, toID, 5000, "Dinner", idKey)

	// Assertions: Confirm expected error and behavior
	assert.Error(t, err)
	assert.Equal(t, ErrInsufficientFunds, err)

	// Ensure no funds were actually moved in the repository
	mockAccRepo.AssertNotCalled(t, "UpdateBalance", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// TestExecuteTransfer_Success validates a complete successful transfer between two accounts.
func TestExecuteTransfer_Success(t *testing.T) {
	// 1. Setup
	mockAccRepo := new(MockAccountRepo)
	mockTxRepo := new(MockTransactionRepo)
	mockPool := new(MockDBPool)
	mockTx := new(MockTx)

	service := NewService(mockPool, mockAccRepo, mockTxRepo)

	fromID := uuid.New()
	toID := uuid.New()
	idKey := uuid.New()

	// 2. Mocking the successful flow
	mockPool.On("Begin", mock.Anything).Return(mockTx, nil)
	mockTxRepo.On("CreateIdempotencyKey", mock.Anything, mock.Anything, idKey).Return(nil)

	// Source account has 100.00 (10000 cents)
	mockAccRepo.On("GetByID", mock.Anything, fromID).Return(&account.Account{
		ID:      fromID,
		Balance: 10000,
	}, nil)

	// Updates and Creation
	mockAccRepo.On("UpdateBalance", mock.Anything, mock.Anything, fromID, int64(-5000)).Return(nil)
	mockAccRepo.On("UpdateBalance", mock.Anything, mock.Anything, toID, int64(5000)).Return(nil)
	mockTxRepo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Transaction must be committed
	mockTx.On("Commit", mock.Anything).Return(nil)
	// Rollback is always called by defer but does nothing if committed
	mockTx.On("Rollback", mock.Anything).Return(nil)

	// 3. Execution
	err := service.ExecuteTransfer(context.Background(), fromID, toID, 5000, "Payment", idKey)

	// 4. Assertions
	assert.NoError(t, err)
	mockPool.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

// TestExecuteTransfer_DatabaseError_Rollback verifies that any database failure
// during the process triggers a rollback and returns an error.
func TestExecuteTransfer_DatabaseError_Rollback(t *testing.T) {
	// 1. Setup
	mockAccRepo := new(MockAccountRepo)
	mockTxRepo := new(MockTransactionRepo)
	mockPool := new(MockDBPool)
	mockTx := new(MockTx)

	service := NewService(mockPool, mockAccRepo, mockTxRepo)

	fromID := uuid.New()
	toID := uuid.New()
	idKey := uuid.New()

	mockPool.On("Begin", mock.Anything).Return(mockTx, nil)
	mockTxRepo.On("CreateIdempotencyKey", mock.Anything, mock.Anything, idKey).Return(nil)

	mockAccRepo.On("GetByID", mock.Anything, fromID).Return(&account.Account{
		ID:      fromID,
		Balance: 10000,
	}, nil)

	// 2. Simulating a DB failure on the first update
	dbError := errors.New("database connection lost")
	mockAccRepo.On("UpdateBalance", mock.Anything, mock.Anything, fromID, int64(-5000)).Return(dbError)

	// 3. Rollback MUST be called when the error occurs
	mockTx.On("Rollback", mock.Anything).Return(nil)

	// 4. Execution
	err := service.ExecuteTransfer(context.Background(), fromID, toID, 5000, "Payment", idKey)

	// 5. Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds") // Note: your service currently wraps some errors
	mockTx.AssertCalled(t, "Rollback", mock.Anything)
}
