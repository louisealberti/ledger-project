// Package api provides the HTTP interface for the Ledger service.
package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"ledger-project/internal/transaction"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// TransferRequest defines the expected JSON payload for a fund transfer.
type TransferRequest struct {
	FromAccountID  uuid.UUID `json:"from_account_id"`
	ToAccountID    uuid.UUID `json:"to_account_id"`
	Amount         int64     `json:"amount"`
	Description    string    `json:"description"`
	IdempotencyKey uuid.UUID `json:"idempotency_key"`
}

// Handler orchestrates HTTP requests and maps them to domain services.
type Handler struct {
	txService *transaction.Service
}

// NewHandler creates a new instance of API Handler with injected dependencies.
func NewHandler(txService *transaction.Service) *Handler {
	return &Handler{
		txService: txService,
	}
}

// TransferHandler handles the POST /transfer request.
// It decodes the JSON body, validates input, and invokes the Transaction Service.
func (h *Handler) TransferHandler(w http.ResponseWriter, r *http.Request) {
	var req TransferRequest

	// 1. Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn().Err(err).Msg("Failed to decode transfer request body")
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 2. Call the Domain Service
	// The service handles atomicity, balance checks, and idempotency.
	err := h.txService.ExecuteTransfer(
		r.Context(),
		req.FromAccountID,
		req.ToAccountID,
		req.Amount,
		req.Description,
		req.IdempotencyKey,
	)

	// 3. Handle Domain Errors and Map to HTTP Status Codes
	if err != nil {
		switch {
		case errors.Is(err, transaction.ErrInsufficientFunds):
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		case errors.Is(err, transaction.ErrAccountNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, transaction.ErrInvalidAmount):
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		default:
			log.Error().Err(err).Msg("Internal server error during transfer")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// 4. Success Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Transfer processed successfully"})
}

// GetBalance handles HTTP requests to fetch an account's balance.
// It expects the account ID as a query parameter (e.g., /balance?id=UUID).
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	accountID, err := uuid.Parse(idStr)
	if err != nil {
		log.Warn().Str("id", idStr).Msg("Invalid UUID provided for balance check")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid account id format"})
		return
	}

	balance, err := h.txService.GetBalance(r.Context(), accountID)
	if err != nil {
		log.Error().Err(err).Interface("account_id", accountID).Msg("Failed to retrieve balance")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "account not found"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"account_id": accountID,
		"balance":    balance,
	})
}

// GetStatement handles the HTTP request to show an account's transaction history.
func (h *Handler) GetStatement(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	accountID, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid account id format"})
		return
	}

	transactions, err := h.txService.GetStatement(r.Context(), accountID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "could not retrieve statement"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}
