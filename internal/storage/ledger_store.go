package storage

import (
	"context"

	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/ledger"
)

type LedgerStore interface {
	SaveEntry(ctx context.Context, entry ledger.LedgerEntry) error
}
