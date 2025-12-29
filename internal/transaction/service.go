// Package transaction manages the business logic for financial movements.
package transaction

import (
	"context"
	"fmt"
	"ledger-project/internal/account"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Service coordinates complex operations between accounts and the database.
type Service struct {
	conn        *pgx.Conn
	accountRepo *account.Repository
	txRepo      *Repository
}

// NewService returns a new transaction service instance.
func NewService(conn *pgx.Conn, ar *account.Repository, tr *Repository) *Service {
	return &Service{
		conn:        conn,
		accountRepo: ar,
		txRepo:      tr,
	}
}

// ExecuteTransfer performs a complete monetary movement between two accounts.
// It ensures atomicity by using a database transaction.
func (s *Service) ExecuteTransfer(ctx context.Context, fromID, toID uuid.UUID, amount int64, desc string) error {
	// 1. Iniciar a transação no banco
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// 2. Buscar as contas (Usando o repository injetado na struct)
	from, err := s.accountRepo.GetByID(ctx, fromID)
	if err != nil {
		return err
	}
	to, err := s.accountRepo.GetByID(ctx, toID)
	if err != nil {
		return err
	}

	// 3. Lógica em memória (Deposit/Withdraw)
	if err := from.Withdraw(amount); err != nil {
		return err
	}
	if err := to.Deposit(amount); err != nil {
		return err
	}

	// 4. Persistir novos saldos
	if err := s.accountRepo.UpdateBalance(ctx, from.ID, from.Balance); err != nil {
		return err
	}
	if err := s.accountRepo.UpdateBalance(ctx, to.ID, to.Balance); err != nil {
		return err
	}

	// 5. Salvar registro da transação
	record := Transaction{
		ID:             uuid.New(),
		AccountFromID:  &from.ID,
		AccountToID:    &to.ID,
		Amount:         amount,
		Description:    desc,
		IdempotencyKey: uuid.New(),
		Status:         "completed",
		CreatedAt:      time.Now(),
	}

	if err := s.txRepo.Save(ctx, record); err != nil {
		return err
	}

	return tx.Commit(ctx)
}