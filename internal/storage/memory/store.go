package memory

import (
	"context" // standard Go package for request-scoped context (timeouts, cancellation)
	"sync"    // standard Go package for concurrency primitives like Mutex

	interfaces "github.com/sheikh-saqib/distributed-payments-ledger-system/internal/interfaces" // interface LedgerStore
	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/models"                // domain models: LedgerEntry
)

// MemoryLedgerStore is an in-memory implementation of storage.LedgerStore.
// It stores ledger entries in memory (slice) and is thread-safe for concurrent writes.
type MemoryLedgerStore struct {
	mu           sync.Mutex                    // mutex to protect entries slice from concurrent access
	entries      []models.LedgerEntry          // slice that holds all ledger entries
	transactions map[string]models.Transaction // slice that holds all transaction entries
}

// NewMemoryLedgerStore creates and returns a new MemoryLedgerStore instance
func NewMemoryLedgerStore() *MemoryLedgerStore {
	return &MemoryLedgerStore{
		entries:      make([]models.LedgerEntry, 0),
		transactions: make(map[string]models.Transaction), // initialize an empty slice of Transactions
	}
}

// SaveEntry saves a LedgerEntry to the in-memory slice.
// Implements the LedgerStore interface.
func (m *MemoryLedgerStore) SaveEntry(ctx context.Context, entry models.LedgerEntry) error {

	m.mu.Lock()         // lock the mutex to prevent concurrent writes
	defer m.mu.Unlock() // unlock automatically when function exits (even if error occurs)

	m.entries = append(m.entries, entry) // append the new entry to the slice
	return nil                           // always succeeds in memory, so returns nil
}

// GetEntries returns a copy of all ledger entries stored in memory.
// Useful for testing, debugging, and printing ledger state.
func (m *MemoryLedgerStore) GetLedgerEntries() ([]models.LedgerEntry, error) {

	m.mu.Lock()         // lock to prevent concurrent modification while reading
	defer m.mu.Unlock() // unlock automatically at the end

	// create a new slice to copy entries
	copied := make([]models.LedgerEntry, len(m.entries))
	copy(copied, m.entries) // copy all entries to the new slice
	return copied, nil      // return the copy so external code can't modify internal state
}

func (m *MemoryLedgerStore) GetEntriesByAccount(accountId string) ([]models.LedgerEntry, error) {

	m.mu.Lock()         // lock the mutex to prevent concurrent writes
	defer m.mu.Unlock() // unlock automatically when function exits (even if error occurs)

	var result []models.LedgerEntry

	for _, e := range m.entries {
		if e.AccountID == accountId {
			result = append(result, e)
		}
	}
	return result, nil
}

func (m *MemoryLedgerStore) TransactionExists(idempotencyKey string) (bool, error) {

	m.mu.Lock()         // lock the mutex to prevent concurrent writes
	defer m.mu.Unlock() // unlock automatically when function exits (even if error occurs)
	_, exists := m.transactions[idempotencyKey]
	return exists, nil
}

func (m *MemoryLedgerStore) SaveTransaction(transaction models.Transaction) error {

	m.mu.Lock()         // lock the mutex to prevent concurrent writes
	defer m.mu.Unlock() // unlock automatically when function exits (even if error occurs)

	m.transactions[transaction.IdempotencyKey] = transaction
	return nil
}

// Compile-time check: ensure MemoryLedgerStore implements LedgerStore interface
var _ interfaces.LedgerStore = (*MemoryLedgerStore)(nil)
