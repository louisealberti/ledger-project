// Package database handles the connection and schema management for the PostgreSQL instance.
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations ensures that the database schema is up to date.
// It creates the necessary tables for accounts, transactions, and idempotency control.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	const query = `
	-- Account storage for holding balances
	CREATE TABLE IF NOT EXISTS accounts (
		id UUID PRIMARY KEY,
		owner_id UUID NOT NULL,
		balance BIGINT NOT NULL DEFAULT 0,
		active BOOLEAN NOT NULL DEFAULT TRUE,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	-- Audit trail for all financial movements
	CREATE TABLE IF NOT EXISTS transactions (
		id UUID PRIMARY KEY,
		account_from_id UUID REFERENCES accounts(id),
		account_to_id UUID REFERENCES accounts(id),
		amount BIGINT NOT NULL,
		description TEXT,
		idempotency_key UUID UNIQUE,
		status TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	-- Control table to prevent double-spending and ensure idempotency
	CREATE TABLE IF NOT EXISTS idempotency_keys (
		id_key UUID PRIMARY KEY,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("database: failed to run migrations: %w", err)
	}

	return nil
}
