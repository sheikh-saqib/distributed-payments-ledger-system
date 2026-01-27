package ledger

import (
	"context"
	"errors"
	"sync"

	interfaces "github.com/sheikh-saqib/distributed-payments-ledger-system/internal/interfaces"
	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/models"
	"github.com/shopspring/decimal"
)

// Ledger is the main struct representing our ledger system
// It holds a reference to the storage layer and a mutex for concurrency control
type Ledger struct {
	store interfaces.LedgerStore // Interface to save ledger entries, can be any storage implementation
	muMap map[string]*sync.Mutex //stores the *sync.Mutex for each account in a map
	mapMu sync.Mutex             // protects the muMap itself
}

// NewLedger is a constructor function that creates a new Ledger instance
// We pass in a storage implementation (MemoryLedgerStore, DB, etc.)
func NewLedger(store interfaces.LedgerStore) *Ledger {
	return &Ledger{
		store: store, // Assign the storage implementation to the ledger's store field
		muMap: make(map[string]*sync.Mutex),
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
func (l *Ledger) PostTransaction(ctx context.Context, tx models.Transaction) error {

	// Idempotency check
	exists, err := l.store.TransactionExists(tx.IdempotencyKey)
	if err != nil {
		return err
	}

	if exists {
		return nil
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
		return errors.New("amount must be positive")
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

	if err := l.store.SaveTransaction(tx); err != nil {
		return err
	}
	// Save the debit entry using the LedgerStore interface
	// If saving fails, return the error immediately
	if err := l.store.SaveEntry(ctx, debit); err != nil {
		return err
	}

	// Save the credit entry using the LedgerStore interface
	// If saving fails, return the error immediately
	// Ensures both sides of the transaction are recorded
	if err := l.store.SaveEntry(ctx, credit); err != nil {
		return err
	}

	// If everything succeeded, return nil indicating no error
	return nil
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
