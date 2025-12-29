// Package database provides structural evolution tools for the PostgreSQL schema.
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// RunMigrations executes the DDL (Data Definition Language) scripts to ensure
// the database schema is up to date. It is idempotent by using 'IF NOT EXISTS'.
func RunMigrations(conn *pgx.Conn) error {
	ctx := context.Background()

	// List of SQL commands to run in order
	queries := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id UUID PRIMARY KEY,
			owner_id UUID NOT NULL,
			balance BIGINT NOT NULL DEFAULT 0,
			active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id UUID PRIMARY KEY,
			account_from_id UUID REFERENCES accounts(id),
			account_to_id UUID REFERENCES accounts(id),
			amount BIGINT NOT NULL,
			description TEXT,
			idempotency_key TEXT UNIQUE NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, q := range queries {
		_, err := conn.Exec(ctx, q)
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	fmt.Println("Migrations executed successfully!")
	return nil
}
