package ledger

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	interfaces "github.com/sheikh-saqib/distributed-payments-ledger-system/internal/interfaces"
	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/models"
	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/models/events"
	"github.com/shopspring/decimal"
)

// Ledger is the main struct representing our ledger system
// It holds a reference to the storage layer and a mutex for concurrency control
type Ledger struct {
	store     interfaces.LedgerStore // Interface to save ledger entries, can be any storage implementation
	muMap     map[string]*sync.Mutex //stores the *sync.Mutex for each account in a map
	mapMu     sync.Mutex             // protects the muMap itself
	appLogger *slog.Logger
	publisher interfaces.EventPublisher
}

// NewLedger is a constructor function that creates a new Ledger instance
// We pass in a storage implementation (MemoryLedgerStore, DB, etc.)
func NewLedger(store interfaces.LedgerStore, appLogger *slog.Logger, publisher interfaces.EventPublisher) *Ledger {
	return &Ledger{
		store:     store, // Assign the storage implementation to the ledger's store field
		appLogger: appLogger,
		publisher: publisher,
		muMap:     make(map[string]*sync.Mutex),
	}
}

func (l *Ledger) getAccountLock(accountId string) *sync.Mutex {

	l.mapMu.Lock()
	defer l.mapMu.Unlock()

	if _, exists := l.muMap[accountId]; !exists {
		l.muMap[accountId] = &sync.Mutex{}
	}
	return l.muMap[accountId]
}

// PostTransaction is the core method that processes a transaction
// It converts a Transaction (intent) into two LedgerEntry objects (debit and credit)
// ensuring double-entry accounting, and then saves them to the store
func (l *Ledger) PostTransaction(ctx context.Context, tx models.Transaction) (bool, error) {
	l.appLogger.Info("received transaction request",
		"idempotency_key", tx.IdempotencyKey,
		"from_account", tx.FromAccount,
		"to_account", tx.ToAccount,
		"amount", tx.Amount.String(),
	)
	// Idempotency check
	exists, err := l.store.TransactionExists(tx.IdempotencyKey)
	if err != nil {
		l.appLogger.Error("transaction failed",
			"error", err.Error(),
			"transaction_id", tx.ID,
		)
		return false, err
	}

	if exists {
		return true, nil
	}
	//Get Locks for both accounts
	debitMutex := l.getAccountLock(tx.FromAccount)
	creditMutex := l.getAccountLock(tx.ToAccount)

	// Lock in order to avoid deadlocks
	if tx.FromAccount < tx.ToAccount {
		debitMutex.Lock()
		creditMutex.Lock()
	} else {
		creditMutex.Lock()
		debitMutex.Lock()
	}

	defer debitMutex.Unlock()
	defer creditMutex.Unlock()

	// Basic validation: the transaction amount must be positive
	if tx.Amount.Cmp(decimal.Zero) <= 0 {
		l.appLogger.Error("amount must be positive")
		return false, errors.New("amount must be positive")
	}
	// Create the debit entry (money leaving the sender's account)
	// - ID: unique entry ID based on transaction ID + "-debit"
	// - AccountID: from which account money is taken
	// - Amount: negative because it's a debit
	// - CreatedAt: timestamp of the transaction
	debit := models.LedgerEntry{
		ID:        tx.ID + "-debit",
		AccountID: tx.FromAccount,
		Amount:    tx.Amount.Neg(),
		CreatedAt: tx.CreatedAt,
	}

	// Create the credit entry (money entering the receiver's account)
	// - ID: unique entry ID based on transaction ID + "-credit"
	// - AccountID: account receiving the money
	// - Amount: positive because it's a credit
	// - CreatedAt: timestamp of the transaction
	credit := models.LedgerEntry{
		ID:        tx.ID + "-credit",
		AccountID: tx.ToAccount,
		Amount:    tx.Amount,
		CreatedAt: tx.CreatedAt,
	}
	l.store.SaveTransactionWithEntries(ctx, tx, debit, credit)
	//Kafka Event
	event := events.TransactionCompleted{
		TransactionID: tx.ID,
		FromAccount:   tx.FromAccount,
		ToAccount:     tx.ToAccount,
		Amount:        tx.Amount,
		OccurredAt:    time.Now(),
	}

	if err := l.publisher.Publish("transactions.completed", event); err != nil {
		l.appLogger.Error("failed to publish kafka event",
			"transaction_id", tx.ID,
			"error", err,
		)
	}
	// If everything succeeded, return nil indicating no error
	return false, nil
}

func (l *Ledger) GetBalance(accountId string) (decimal.Decimal, error) {
	ledgerEntries, err := l.store.GetEntriesByAccount(accountId)

	if err != nil {
		return decimal.Zero, err
	}
	balance := decimal.Zero

	for _, ledgerEntry := range ledgerEntries {
		balance = balance.Add(ledgerEntry.Amount)
	}
	return balance, nil
}
func (l *Ledger) GetLedgerEntries() ([]models.LedgerEntry, error) {
	ledgerEntries, err := l.store.GetLedgerEntries()

	if err != nil {
		return []models.LedgerEntry{}, err
	}
	return ledgerEntries, nil
}
