package interfaces

import (
	"context"
	"database/sql"

	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/models"
)

type LedgerStore interface {
	SaveTransactionWithEntries(ctx context.Context, tx models.Transaction, debit models.LedgerEntry, credit models.LedgerEntry) error
	GetEntriesByAccount(accountId string) ([]models.LedgerEntry, error)
	GetLedgerEntries() ([]models.LedgerEntry, error)

	TransactionExists(idempotencyKey string) (bool, error)
	SaveTransaction(tx models.Transaction, dbTx *sql.Tx) error
}
