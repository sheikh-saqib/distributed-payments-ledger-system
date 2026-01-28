CREATE TABLE ledger_entries (
    id TEXT PRIMARY KEY,           -- Unique ledger entry ID
    account_id TEXT NOT NULL,      -- Which account this entry belongs to
    amount NUMERIC(20,8) NOT NULL,-- Amount (decimal, positive or negative)
    created_at TIMESTAMP NOT NULL  -- Timestamp of the entry
);

-- Index to make balance queries fast
CREATE INDEX idx_ledger_entries_account_id
ON ledger_entries(account_id);


CREATE TABLE transactions (
    id TEXT PRIMARY KEY,               -- Logical transaction ID
    idempotency_key TEXT NOT NULL UNIQUE, -- Prevent duplicate processing
    from_account TEXT NOT NULL,        -- Sender
    to_account TEXT NOT NULL,          -- Receiver
    amount NUMERIC(20,8) NOT NULL,    -- Transaction amount
    created_at TIMESTAMP NOT NULL      -- Timestamp of the transaction
);
