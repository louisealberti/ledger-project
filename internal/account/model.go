// Package account contains the core domain logic for banking operations.
package account

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Account represents a financial entity within the system.
type Account struct {
	// Unique identifier for the account (UUID)
	ID        uuid.UUID `json:"id"`
	Balance   int64     `json:"balance"`
	Active    bool      `json:"active"`
	UpdatedAt time.Time `json:"updated_at"`
	OwnerID   uuid.UUID `json:"owner_id"`
}

// IsAvailable checks if the account is in a valid state for operations.
func (a *Account) IsAvailable() error {

	if !a.Active {
		return errors.New("account is inactive")
	}
	return nil
}

// CanWithdraw validates both account availability and sufficient funds.
func (a *Account) CanWithdraw(amount int64) error {

	if err := a.IsAvailable(); err != nil {
		return err
	}
	if a.Balance < amount {
		return errors.New("insufficient balance")
	}
	return nil
}

// Deposit increases the account balance after validating its availability.
func (a *Account) Deposit(amount int64) error {
	if err := a.IsAvailable(); err != nil {
		return err
	}
	a.Balance += amount
	a.UpdatedAt = time.Now()
	return nil
}

// Withdraw decreases the account balance after validating availability and sufficient funds.
func (a *Account) Withdraw(amount int64) error {

	err := a.CanWithdraw(amount)
	if err != nil {
		return err // If validation fails, we stop here and return the error
	}

	a.Balance -= amount
	a.UpdatedAt = time.Now()

	return nil
}
