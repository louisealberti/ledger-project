// Package api provides the HTTP interface for interacting with the Ledger system.
package api

import (
	"encoding/json"
	"ledger-project/internal/transaction"
	"net/http"

	"github.com/google/uuid"
)

// TransferRequest defines the expected JSON payload for a fund movement.
// Account IDs are received as strings to allow validation during the parsing layer.
type TransferRequest struct {
	FromID      string `json:"account_from_id"`
	ToID        string `json:"account_to_id"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

// Handler orchestrates HTTP requests and routes them to the appropriate services.
type Handler struct {
	txService *transaction.Service
}

// NewHandler returns a new API handler with the required service dependencies.
func NewHandler(s *transaction.Service) *Handler {
	return &Handler{txService: s}
}

// TransferHandle processes an incoming transfer request via POST.
// It decodes the JSON body, validates the UUID format, and invokes the transaction service.
func (h *Handler) TransferHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Input Validation: Ensure IDs are valid UUIDs
	fromUUID, err := uuid.Parse(req.FromID)
	if err != nil {
		http.Error(w, "Invalid account_from_id format", http.StatusBadRequest)
		return
	}

	toUUID, err := uuid.Parse(req.ToID)
	if err != nil {
		http.Error(w, "Invalid account_to_id format", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err = h.txService.ExecuteTransfer(ctx, fromUUID, toUUID, req.Amount, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status": "transfer completed successfully"}`))
}