package postgres

import (
	"context"
	"database/sql"

	interfaces "github.com/sheikh-saqib/distributed-payments-ledger-system/internal/interfaces" // interface LedgerStore
	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/models"
)

type PostgresLedgerStore struct {
	db *sql.DB
}

func NewPostgresLedgerStore(db *sql.DB) *PostgresLedgerStore {
	return &PostgresLedgerStore{
		db: db,
	}
}

func (p *PostgresLedgerStore) TransactionExists(idempotencyKey string) (bool, error) {
	const query = `select 1 from transactions where idempotency_key = $1 Limit 1`

	var exists int
	err := p.db.QueryRow(query, idempotencyKey).Scan(&exists)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (p *PostgresLedgerStore) SaveTransaction(tx models.Transaction, dbTx *sql.Tx) error {
	const query = `INSERT INTO transactions(id, idempotency_key,from_account,to_account,amount,created_at)
	VALUES ($1,$2,$3,$4,$5,$6)`

	_, err := dbTx.Exec(query, tx.ID, tx.IdempotencyKey, tx.FromAccount, tx.ToAccount, tx.Amount, tx.CreatedAt)

	return err
}

func (p *PostgresLedgerStore) SaveEntry(ctx context.Context, ledgerEntry models.LedgerEntry, dbTx *sql.Tx) error {
	const query = `INSERT INTO ledger_entries (id,account_id, amount,created_at)
	VALUES ($1,$2,$3,$4)`

	_, err := dbTx.ExecContext(ctx, query, ledgerEntry.ID, ledgerEntry.AccountID, ledgerEntry.Amount, ledgerEntry.CreatedAt)
	return err
}

func (p *PostgresLedgerStore) SaveTransactionWithEntries(ctx context.Context, tx models.Transaction, debit models.LedgerEntry, credit models.LedgerEntry) error {

	dbTx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			dbTx.Rollback()
		}
	}()

	err = p.SaveTransaction(tx, dbTx)
	if err != nil {
		return err
	}

	err = p.SaveEntry(ctx, debit, dbTx)
	if err != nil {
		return err
	}

	err = p.SaveEntry(ctx, credit, dbTx)
	if err != nil {
		return err
	}
	return dbTx.Commit()
}

func (p *PostgresLedgerStore) GetLedgerEntries() ([]models.LedgerEntry, error) {

	const query = `SELECT id, account_id, amount, created_at from ledger_entries`

	rows, err := p.db.Query(query)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var entries []models.LedgerEntry

	for rows.Next() {
		var entry models.LedgerEntry
		err := rows.Scan(
			&entry.ID,
			&entry.AccountID,
			&entry.Amount,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func (p *PostgresLedgerStore) GetEntriesByAccount(accountId string) ([]models.LedgerEntry, error) {
	const query = `SELECT id, account_id, amount, created_at from ledger_entries 
	WHERE account_id = $1`

	rows, err := p.db.Query(query, accountId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var entries []models.LedgerEntry
	for rows.Next() {
		var entry models.LedgerEntry
		if err := rows.Scan(&entry.ID, &entry.AccountID, &entry.Amount, &entry.CreatedAt); err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}
	return entries, nil
}

var _ interfaces.LedgerStore = (*PostgresLedgerStore)(nil)
