// Package transaction manages the immutable records of monetary movements.
package transaction

import (
	"time"

	"github.com/google/uuid"
)

// Transaction represents a single financial movement between two accounts.
// It follows the double-entry bookkeeping principle.
type Transaction struct {
	ID             uuid.UUID  // Unique UUID for the transaction
	AccountFromID  *uuid.UUID // Nullable for deposits (money entering the system)
	AccountToID    *uuid.UUID // Nullable for withdrawals (money leaving the system)
	Amount         int64      // Value in cents
	Description    string     // Context for the movement
	IdempotencyKey uuid.UUID  // Guard against duplicate processing
	Status         string     // e.g., "completed", "reversed", "failed"
	CreatedAt      time.Time  // Timestamp of the event
}
