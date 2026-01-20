package ledger

import "time"

// LedgerEntry represents a single ledger record for an account
type LedgerEntry struct {
	ID        string    // unique identifier
	AccountID string    // which account this entry belongs to
	Amount    int64     // in cents (positive or negative)
	CreatedAt time.Time // timestamp
}
