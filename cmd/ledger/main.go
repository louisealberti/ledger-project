// Package main is the entry point of the Ledger service.
package main

import (
	"context"
	"fmt"
	"ledger-project/internal/account"
	"ledger-project/internal/api"
	"ledger-project/internal/database"
	"ledger-project/internal/transaction"
	"net/http"
)

func main() {
	ctx := context.Background()

	// 1. Infra: Database Connection
	conn := database.ConnectDB()
	defer conn.Close(ctx)

	// 2. Infra: Database Migrations
	if err := database.RunMigrations(conn); err != nil {
		fmt.Printf("Startup error (Migrations): %v\n", err)
		return
	}

	// 2.1 Populates database with test data
	if err := SeedInitialData(ctx, conn); err != nil {
		fmt.Printf("Startup error (Seed): %v\n", err)
	}

	// 3. Domain: Repositories
	accRepo := account.NewRepository(conn)
	txRepo := transaction.NewRepository(conn)

	// 4. Domain: Services (The logic engine)
	txService := transaction.NewService(conn, accRepo, txRepo)

	// 5. Interface: API Handlers and Router
	handler := api.NewHandler(txService)
	router := api.NewRouter(handler)

	// 6. Starting the Server
	port := ":8080"
	fmt.Printf("Ledger API online at http://localhost%s\n", port)

	if err := http.ListenAndServe(port, router); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
