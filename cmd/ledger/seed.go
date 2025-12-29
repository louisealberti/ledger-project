// Package main provides administrative utilities for the Ledger system.
package main

import (
	"context"
	"fmt"
	"ledger-project/internal/account"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// SeedInitialData populates the database with test accounts to facilitate development.
func SeedInitialData(ctx context.Context, conn *pgx.Conn) error {
	// Removi a linha 'repo := account.NewRepository(conn)' pois não estava sendo usada
	// e o Go proíbe variáveis declaradas e não utilizadas.

	// Test Account 1: The Sender (1000.00 BRL)
	acc1 := account.Account{
		ID:        uuid.New(), // Removido .String()
		OwnerID:   uuid.New(), // Removido .String()
		Balance:   100000,
		Active:    true,
		UpdatedAt: time.Now(),
	}

	// Test Account 2: The Receiver (0.00 BRL)
	acc2 := account.Account{
		ID:        uuid.New(), // Removido .String()
		OwnerID:   uuid.New(), // Removido .String()
		Balance:   0,
		Active:    true,
		UpdatedAt: time.Now(),
	}

	for _, acc := range []account.Account{acc1, acc2} {
		query := `INSERT INTO accounts (id, owner_id, balance, active, updated_at) 
		          VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`

		_, err := conn.Exec(ctx, query, acc.ID, acc.OwnerID, acc.Balance, acc.Active, acc.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to seed account %s: %w", acc.ID, err)
		}
		fmt.Printf("Account seeded: %s | Balance: %d\n", acc.ID, acc.Balance)
	}

	return nil
}
