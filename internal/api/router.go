package api

import (
	"net/http"
)

// NewRouter initializes the API routes and returns an http.Handler.
// We use the standard library here, but this could easily be swapped for Echo/Gin.
func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	// Route definitions
	mux.HandleFunc("POST /transfer", h.TransferHandler)
	mux.HandleFunc("GET /balance", h.GetBalance)
	mux.HandleFunc("GET /statement", h.GetStatement)

	return mux
}
