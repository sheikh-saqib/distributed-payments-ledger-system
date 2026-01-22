package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// LedgerEntry represents a single ledger record for an account
type LedgerEntry struct {
	ID        string          // unique identifier
	AccountID string          // which account this entry belongs to
	Amount    decimal.Decimal // in cents (positive or negative)
	CreatedAt time.Time       // timestamp
}
