// Package account handles the persistence logic for bank accounts.
package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// SQL to create a new account
	createAccountQuery = `INSERT INTO accounts (id, owner_id, balance, active) VALUES ($1, $2, $3, $4)`

	// SQL to fetch an account by its unique identifier
	getQuery = `SELECT id, owner_id, balance, active FROM accounts WHERE id = $1`

	// SQL to update balance with an atomic safety check
	updateBalanceQuery = `
		UPDATE accounts 
		SET balance = balance + $1 
		WHERE id = $2 AND (balance + $1) >= 0`
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new account record into the database.
func (r *Repository) Create(ctx context.Context, acc *Account) error {
	_, err := r.pool.Exec(ctx, createAccountQuery, acc.ID, acc.OwnerID, acc.Balance, acc.Active)
	if err != nil {
		return fmt.Errorf("repository: failed to create account: %w", err)
	}
	return nil
}

// GetByID retrieves an account from the database.
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Account, error) {
	var acc Account
	err := r.pool.QueryRow(ctx, getQuery, id).Scan(
		&acc.ID,
		&acc.OwnerID,
		&acc.Balance,
		&acc.Active,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("account not found")
		}
		return nil, err
	}

	return &acc, nil
}

// UpdateBalance applies a delta to the account balance.
// It returns an error if the operation would result in a negative balance.
func (r *Repository) UpdateBalance(ctx context.Context, tx pgx.Tx, id uuid.UUID, amount int64) error {
	result, err := tx.Exec(ctx, updateBalanceQuery, amount, id)
	if err != nil {
		return fmt.Errorf("repository: failed to update balance: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("repository: balance update failed (insufficient funds or invalid account)")
	}

	return nil
}
