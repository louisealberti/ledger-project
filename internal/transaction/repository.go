// Package transaction provides persistence logic for financial movements.
package transaction

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Repository handles all direct communication with the transactions table.
type Repository struct {
	conn *pgx.Conn
}

// NewRepository returns a new instance of the transaction repository.
func NewRepository(conn *pgx.Conn) *Repository {
	return &Repository{conn: conn}
}

// Save records a new transaction entry in the database.
func (r *Repository) Save(ctx context.Context, tx Transaction) error {
	const query = `
		INSERT INTO transactions (
			id, account_from_id, account_to_id, amount, description, idempotency_key, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.conn.Exec(ctx, query,
		tx.ID,
		tx.AccountFromID,
		tx.AccountToID,
		tx.Amount,
		tx.Description,
		tx.IdempotencyKey,
		tx.Status,
		tx.CreatedAt,
	)

	return err
}
