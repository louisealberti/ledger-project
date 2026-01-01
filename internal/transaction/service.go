// Package transaction implements the core business rules for ledger operations.
package transaction

import (
	"context"
	"fmt"
	"ledger-project/internal/account"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"     // used for the Logger type
	"github.com/rs/zerolog/log" // used for the global logger instance
)

type AccountRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*account.Account, error)
	UpdateBalance(ctx context.Context, tx pgx.Tx, id uuid.UUID, amount int64) error
}

type TransactionRepository interface {
    CreateIdempotencyKey(ctx context.Context, tx pgx.Tx, key uuid.UUID) error
    Create(ctx context.Context, tx pgx.Tx, t *Transaction) error
    GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]Transaction, error)
}

// DBPool defines the behavior required to start a database transaction.
// This allows us to swap the real pgxpool.Pool for a Mock during unit tests.
type DBPool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Service struct {
	pool    DBPool
	accRepo AccountRepository
	txRepo  TransactionRepository
}

func NewService(pool DBPool, accRepo AccountRepository, txRepo TransactionRepository) *Service {
	return &Service{pool: pool, accRepo: accRepo, txRepo: txRepo}
}

// ExecuteTransfer orchestrates the fund transfer between two accounts within a database transaction.
// It ensures atomicity, idempotency, and provides structured logging for observability.
func (s *Service) ExecuteTransfer(ctx context.Context, fromID, toID uuid.UUID, amount int64, desc string, idKey uuid.UUID) error {
	// Creating a sub-logger with the Idempotency Key to trace this specific request
	subLog := log.With().
		Str("component", "transaction_service").
		Interface("idKey", idKey).
		Logger()

	subLog.Info().
		Interface("from_id", fromID).
		Interface("to_id", toID).
		Int64("amount", amount).
		Msg("Processing transfer request")

	if amount <= 0 {
		subLog.Warn().Int64("amount", amount).Msg("Transfer rejected: invalid amount")
		return ErrInvalidAmount
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		subLog.Error().Err(err).Msg("Database connection error: failed to begin transaction")
		return err
	}
	defer tx.Rollback(ctx)

	// Passing the sub-logger to the internal process to maintain context
	if err := s.processTransfer(ctx, tx, fromID, toID, amount, desc, idKey, subLog); err != nil {
		if err.Error() == "IDEMPOTENCY_STOP" {
			subLog.Info().Msg("Request skipped: idempotency key already processed")
			return nil
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		subLog.Error().Err(err).Msg("Critical failure: transaction commit failed")
		return err
	}

	subLog.Info().Msg("Transfer successfully executed")
	return nil
}

// processTransfer contains the core business logic.
// It uses zerolog.Logger as the type for the logger parameter.
func (s *Service) processTransfer(ctx context.Context, tx pgx.Tx, fromID, toID uuid.UUID, amount int64, desc string, idKey uuid.UUID, logger zerolog.Logger) error {
	// 1. Idempotency Check
	if err := s.txRepo.CreateIdempotencyKey(ctx, tx, idKey); err != nil {
		return fmt.Errorf("IDEMPOTENCY_STOP")
	}

	// 2. Source Account Validation
	fromAcc, err := s.accRepo.GetByID(ctx, fromID)
	if err != nil {
		logger.Error().Err(err).Interface("account_id", fromID).Msg("Source account lookup failed")
		return ErrAccountNotFound
	}

	if fromAcc.Balance < amount {
		logger.Warn().
			Int64("current_balance", fromAcc.Balance).
			Int64("requested_amount", amount).
			Msg("Transfer denied: insufficient balance")
		return ErrInsufficientFunds
	}

	// 3. Atomic Balance Updates
	if err := s.accRepo.UpdateBalance(ctx, tx, fromID, -amount); err != nil {
		// This should only fail if balance changed between GetByID and UpdateBalance
		return ErrInsufficientFunds
	}

	if err := s.accRepo.UpdateBalance(ctx, tx, toID, amount); err != nil {
		logger.Error().Err(err).Interface("to_id", toID).Msg("Failed to credit destination account")
		return fmt.Errorf("service: credit failed: %w", err)
	}

	// 4. Record Transaction History
	newTx := Transaction{
		ID:             uuid.New(),
		AccountFromID:  &fromID,
		AccountToID:    &toID,
		Amount:         amount,
		Description:    desc,
		IdempotencyKey: idKey,
		Status:         "completed",
		CreatedAt:      time.Now(),
	}

	return s.txRepo.Create(ctx, tx, &newTx)
}

// GetBalance retrieves the current balance of a specific account.
// It returns the balance in cents (int64) to avoid floating point issues.
func (s *Service) GetBalance(ctx context.Context, accountID uuid.UUID) (int64, error) {
	acc, err := s.accRepo.GetByID(ctx, accountID)
	if err != nil {
		return 0, err
	}
	return acc.Balance, nil
}

// GetStatement retrieves all credit and debit transactions for a specific account.
// It returns a slice of Transaction models or an error if the account cannot be accessed.
func (s *Service) GetStatement(ctx context.Context, accountID uuid.UUID) ([]Transaction, error) {

	transactions, err := s.txRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		log.Error().Err(err).Interface("account_id", accountID).Msg("Failed to fetch transactions for statement")
		return nil, err
	}

	return transactions, nil
}
