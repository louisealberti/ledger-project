// Package api handles the routing logic for the ledger server.
package api

import "net/http"

// NewRouter sets up the routes and returns a standard http.Handler.
func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	// Registering the transfer endpoint
	mux.HandleFunc("/transfer", h.TransferHandle)

	return mux
}
