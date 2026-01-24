package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Transaction represents an intent to transfer money
type Transaction struct {
	ID             string
	IdempotencyKey string
	FromAccount    string
	ToAccount      string
	Amount         decimal.Decimal
	CreatedAt      time.Time
	Replayed       bool
}
