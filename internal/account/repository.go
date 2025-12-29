// Package account provides persistence logic for banking entities.
package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Repository handles all direct communication with the accounts table.
type Repository struct {
	conn *pgx.Conn
}

// NewRepository returns a new instance of the account repository.
func NewRepository(conn *pgx.Conn) *Repository {
	return &Repository{conn: conn}
}

// GetByID retrieves an account from the database by its unique identifier.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	const query = `SELECT id, owner_id, balance, active, updated_at FROM accounts WHERE id = $1`

	var acc Account
	err := r.conn.QueryRow(ctx, query, id).Scan(
		&acc.ID, &acc.OwnerID, &acc.Balance, &acc.Active, &acc.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &acc, nil
}

// UpdateBalance updates the balance and updated_at timestamp of a specific account.
func (r *Repository) UpdateBalance(ctx context.Context, id uuid.UUID, newBalance int64) error {
	const query = `UPDATE accounts SET balance = $1, updated_at = NOW() WHERE id = $2`

	_, err := r.conn.Exec(ctx, query, newBalance, id)
	return err
}
