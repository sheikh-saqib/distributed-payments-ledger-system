package interfaces

import (
	"context"

	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/models"
)

type LedgerStore interface {
	SaveEntry(ctx context.Context, entry models.LedgerEntry) error
	GetEntriesByAccount(accountId string) ([]models.LedgerEntry, error)
	GetLedgerEntries() ([]models.LedgerEntry, error)

	TransactionExists(idempotencyKey string) (bool, error)
	SaveTransaction(tx models.Transaction) error
}
