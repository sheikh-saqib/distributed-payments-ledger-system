package ledger

import (
	"context"
	"errors"
	"sync"

	interfaces "github.com/sheikh-saqib/distributed-payments-ledger-system/internal/interfaces"
	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/models"
)

// Ledger is the main struct representing our ledger system
// It holds a reference to the storage layer and a mutex for concurrency control
type Ledger struct {
	store interfaces.LedgerStore // Interface to save ledger entries, can be any storage implementation
	mu    sync.Mutex             // Mutex ensures thread-safe writes to prevent race conditions
}

// NewLedger is a constructor function that creates a new Ledger instance
// We pass in a storage implementation (MemoryLedgerStore, DB, etc.)
func NewLedger(store interfaces.LedgerStore) *Ledger {
	return &Ledger{
		store: store, // Assign the storage implementation to the ledger's store field
	}
}

// PostTransaction is the core method that processes a transaction
// It converts a Transaction (intent) into two LedgerEntry objects (debit and credit)
// ensuring double-entry accounting, and then saves them to the store
func (l *Ledger) PostTransaction(ctx context.Context, tx models.Transaction) error {

	// Acquire the mutex lock before entering the critical section
	// This ensures that only one transaction is processed at a time
	l.mu.Lock()
	// defer ensures that the lock will be released when the function exits
	// even if an error occurs or the function returns early
	defer l.mu.Unlock()

	// Basic validation: the transaction amount must be positive
	if tx.Amount <= 0 {
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
		Amount:    -tx.Amount,
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
