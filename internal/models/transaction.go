package models

import "time"

// Transaction represents an intent to transfer money
type Transaction struct {
	ID          string
	FromAccount string
	ToAccount   string
	Amount      float64
	CreatedAt   time.Time
}
