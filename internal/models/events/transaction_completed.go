package events

import (
	"time"

	"github.com/shopspring/decimal"
)

type TransactionCompleted struct {
	TransactionID string          `json:"transaction_id"`
	FromAccount   string          `json:"from_account"`
	ToAccount     string          `json:"to_account"`
	Amount        decimal.Decimal `json:"amount"`
	OccurredAt    time.Time       `json:"occurred_at"`
}
