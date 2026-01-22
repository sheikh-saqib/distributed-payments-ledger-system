# Architecture Decisions

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
**Decision**: Single `sync.RWMutex` serializes all writes.

**Why**: Correctness first. No race conditions possible.

**Trade-off**: 
- ✅ Simple and correct
- ❌ Bottleneck—blocks ALL accounts even if unrelated
- 🔜 **Phase 3 will fix this** with per-account locking

This is **intentional technical debt**.

---

### 4. Money as int64 (Cents)
**Decision**: Store amounts as smallest unit (e.g., 1000 = £10.00).

**Why**: No floating-point precision errors. Industry standard (Stripe, Monzo).

**Trade-off**: Must convert for display, but financial correctness matters more.

---

### 5. Thin HTTP Handlers
**Decision**: Handlers only validate input and call ledger service.

**Why**: Business logic stays testable. Can add gRPC/CLI without duplication.

**Trade-off**: More files, but better architecture.

---

## Known Limitations (Current State)

- ❌ Global mutex blocks unrelated accounts → **Phase 3**
- ❌ No balance queries → **Phase 4**
- ❌ No idempotency (duplicate protection) → **Phase 4**
- ❌ Data lost on restart → **Phase 5**

These are deliberate. Each phase solves one problem well.

---

## Principles

1. Correctness over performance (for now)
2. Simplicity over cleverness
3. Interfaces over implementations
4. Progress over perfection