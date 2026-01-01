// Package transaction provides the persistence layer for ledger movements.
// This file handles SQL execution for transactions and idempotency controls.
package transaction

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL queries defined as constants to keep methods clean and readable.
const (
	queryInsertIdempotency = `INSERT INTO idempotency_keys (id_key, response_status, response_body) VALUES ($1, $2, $3)`
	queryInsertTransaction = `INSERT INTO transactions (id, account_from_id, account_to_id, amount, description, idempotency_key, status, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	queryGetTransactions   = `SELECT id, account_from_id, account_to_id, amount, description, idempotency_key, status, created_at FROM transactions WHERE account_from_id = $1 OR account_to_id = $1 ORDER BY created_at DESC`
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateIdempotencyKey(ctx context.Context, tx pgx.Tx, key uuid.UUID) error {
	_, err := tx.Exec(ctx, queryInsertIdempotency, key, 201, "success")
	return err
}

func (r *Repository) Create(ctx context.Context, tx pgx.Tx, t *Transaction) error {
	_, err := tx.Exec(ctx, queryInsertTransaction,
		t.ID, t.AccountFromID, t.AccountToID, t.Amount, t.Description, t.IdempotencyKey, t.Status, t.CreatedAt)
	return err
}

// GetByAccountID retrieves all transactions related to a specific account.
func (r *Repository) GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]Transaction, error) {
	var transactions []Transaction

	rows, err := r.pool.Query(ctx, queryGetTransactions, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t Transaction
		err := rows.Scan(
			&t.ID, &t.AccountFromID, &t.AccountToID, &t.Amount,
			&t.Description, &t.IdempotencyKey, &t.Status, &t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	return transactions, nil
}
