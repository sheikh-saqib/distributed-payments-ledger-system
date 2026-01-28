# Architecture Decisions

---

## Phase 2: Ledger & API

### 1. Ledger-Based Accounting (Not CRUD)

**Decision**: Double-entry bookkeeping with immutable, append-only entries.

**Why**: This is how real banks work. Every transaction creates a debit and credit entry that can never be changed.

**Trade-off**: More storage, but complete audit trail and no data corruption.

---

### 2. Interface-Driven Storage

**Decision**: `LedgerStore` interface with in-memory implementation first.

**Why**: Can test business logic without a database. Easy to swap storage later (Postgres, Cassandra).

**Trade-off**: Slight abstraction overhead, but clean separation and testability.

---

### 3. Global Mutex (Temporary)

**Decision**: Single `sync.Mutex` serializes all writes.

**Why**: Correctness first. No race conditions possible.

**Trade-off**:

* ✅ Simple and correct
* ❌ Bottleneck—blocks ALL accounts even if unrelated
* 🔜 **Phase 3 will fix this**

This is **intentional technical debt**.

---

### 4. Money Representation (Decimal)

**Decision**: Use `decimal.Decimal` for monetary values.

**Why**:

* Avoids floating-point precision errors
* Supports arbitrary precision
* Safe for financial calculations

**Trade-off**: Slight performance overhead vs `int64`, but correctness is more important.

---

### 5. Thin HTTP Handlers

**Decision**: HTTP handlers only validate input and call ledger service.

**Why**: Business logic stays testable. Can add gRPC/CLI without duplication.

**Trade-off**: More files, but cleaner architecture.

---

## Phase 3: Concurrency & Correctness (Implemented)

### 6. Per-Account Locking

**Decision**: Replace the global mutex with **per-account mutexes** stored in a map.

```go
muMap map[string]*sync.Mutex
```

**Implementation**:

* Each account ID maps to a `*sync.Mutex`
* `getAccountLock(accountID)` retrieves or creates the mutex
* Ensures transactions touching the same account are serialized

**Why**:

* Enables concurrent transactions on unrelated accounts
* Prevents race conditions for the same account

---

### 7. Deadlock Prevention

**Decision**: Lock accounts in deterministic order (by account ID).

**Implementation**:

```go
if tx.FromAccount < tx.ToAccount {
    debitMutex.Lock()
    creditMutex.Lock()
} else {
    creditMutex.Lock()
    debitMutex.Lock()
}
defer debitMutex.Unlock()
defer creditMutex.Unlock()
```

**Why**:

* Prevents circular wait / deadlocks
* Ensures system safety under concurrency

---

### 8. Separate Mutex for Lock Map (`mapMu`)

**Decision**: Protect the `muMap` itself with a separate mutex.

**Implementation**:

* `mapMu sync.Mutex` guards creation/retrieval of account locks
* Ensures only one mutex exists per account

---

## Phase 4: Queries & Idempotency (Implemented)

### 9. Balance Computation

**Decision**: Account balances are computed by summing all ledger entries for that account.

**Implementation**:

* `GetBalance(accountID)` iterates ledger entries and sums amounts
* Returns `decimal.Decimal` balance
* No balance field is stored — derived from entries

**Why**:

* Prevents balance inconsistency
* Matches event-sourced systems

---

### 10. Idempotent Transactions

**Decision**: Use `IdempotencyKey` to prevent duplicate processing.

**Implementation**:

* `TransactionExists(idempotencyKey)` checks in-memory slice
* `SaveTransaction(tx)` stores transactions separately from ledger entries
* `PostTransaction` first checks idempotency before creating entries

**Why**:

* Ensures at-most-once transaction semantics
* Protects against retries

---

### 11. Separate Storage for Transactions vs Ledger Entries

**Decision**: Transactions represent intent; ledger entries represent accounting facts.

**Implementation**:

* `MemoryLedgerStore.entries []LedgerEntry`
* `MemoryLedgerStore.transactions []Transaction`

**Why**:

* Separation of concerns keeps logic clear and testable
* Transactions are for idempotency; entries are for auditing

---

## Phase 5: Persistence & Atomicity (Implemented)

### 12. PostgreSQL Persistence

**Decision**: Persist ledger and transaction data in Postgres for durability.

**Implementation**:

* `PostgresLedgerStore` implements `LedgerStore` interface
* Stores `ledger_entries` and `transactions` tables
* Queries and writes use `*sql.DB` and `*sql.Tx` for atomic operations

**Why**:

* Data is persistent across restarts
* Idempotency lookup is fast and reliable
* Supports real concurrency

---

### 13. Atomic Transactions

**Decision**: Use SQL transactions (`BEGIN/COMMIT/ROLLBACK`) to ensure all-or-nothing writes.

**Implementation**:

* `SaveTransactionWithEntries(ctx, tx, debit, credit)` wraps inserts in `db.BeginTx`
* On any failure, `Rollback()` ensures no partial writes
* On success, `Commit()` saves all entries atomically

**Why**:

* Guarantees ledger consistency
* Prevents partial debits/credits
* Matches production banking standards

---

### 14. Idempotency at DB Level

**Decision**: Transactions are only inserted if `idempotency_key` does not exist.

**Implementation**:

* `TransactionExists(idempotencyKey)` queries DB first
* Prevents double-posting in concurrent requests

**Why**:

* Ensures at-most-once semantics even across server restarts
* Protects against retry storms

---

## Known Limitations

* ❌ No database indexes yet → may slow queries for large datasets
* ❌ No migrations system → schema changes are manual
* ❌ No pagination or historical queries → future work
* ❌ No metrics or monitoring → future work

These limitations are **intentional and phased**.

---

## Principles

1. Correctness over performance
2. Explicit concurrency over hidden magic
3. Ledger entries over mutable balances
4. Interfaces over implementations
5. Build like a bank, not a CRUD app
6. Always prefer durability and atomicity over shortcuts
