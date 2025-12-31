// Package main provides administrative utilities for the Ledger system.
package main

import (
	"context"
	"fmt"
	"ledger-project/internal/account"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log" // Importando o logger global
)

// SeedInitialData populates the database with test accounts to facilitate development.
// It uses pgxpool.Pool for concurrent-safe database access.
func SeedInitialData(ctx context.Context, db *pgxpool.Pool) error {
	// Creating fixed UUIDs for development makes testing with CURL much easier.
	// However, if you prefer dynamic ones, we will log them clearly.

	acc1 := account.Account{
		ID:        uuid.New(),
		OwnerID:   uuid.New(),
		Balance:   100000,
		Active:    true,
		UpdatedAt: time.Now(),
	}

	acc2 := account.Account{
		ID:        uuid.New(),
		OwnerID:   uuid.New(),
		Balance:   0,
		Active:    true,
		UpdatedAt: time.Now(),
	}

	for _, acc := range []account.Account{acc1, acc2} {
		query := `INSERT INTO accounts (id, owner_id, balance, active, updated_at) 
                  VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`

		_, err := db.Exec(ctx, query, acc.ID, acc.OwnerID, acc.Balance, acc.Active, acc.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to seed account %s: %w", acc.ID, err)
		}

		// Usando Zerolog para manter o padrão profissional do projeto
		log.Info().
			Interface("account_id", acc.ID).
			Int64("balance", acc.Balance).
			Msg("Database record seeded successfully")
	}

	return nil
}
