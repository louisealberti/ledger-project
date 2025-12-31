// Package transaction defines domain-specific errors for financial operations.
// These sentinel errors allow the application to distinguish between
// business rule violations and system failures.
package transaction

import "errors"

var (
	// ErrInsufficientFunds is returned when the source account balance is lower than the transfer amount.
	ErrInsufficientFunds = errors.New("insufficient funds")

	// ErrInvalidAmount is returned when the transfer amount is zero or negative.
	ErrInvalidAmount = errors.New("invalid transfer amount")

	// ErrAccountNotFound is returned when either the source or destination account does not exist.
	ErrAccountNotFound = errors.New("account not found")
)
