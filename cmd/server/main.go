package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	interfaces "github.com/sheikh-saqib/distributed-payments-ledger-system/internal/interfaces"
	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/ledger"
	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/models"

	// "github.com/sheikh-saqib/distributed-payments-ledger-system/internal/storage/memory"
	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/storage/postgres"
	"github.com/shopspring/decimal"
)

func main() {

	// var store interfaces.LedgerStore = memory.NewMemoryLedgerStore()
	// ledgerService := ledger.NewLedger(store)
	// PostgreSQL connection string

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found.")
	}

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	// Ping to check connection
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	// Inject DB into PostgresLedgerStore
	var store interfaces.LedgerStore = postgres.NewPostgresLedgerStore(db)

	// Create Ledger service with Postgres store
	ledgerService := ledger.NewLedger(store)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// 3️⃣ Transactions endpoint (NEW)
	http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		idempotencyKey := r.Header.Get("Idempotency-Key")

		var req struct {
			FromAccount string          `json:"from_account"`
			ToAccount   string          `json:"to_account"`
			Amount      decimal.Decimal `json:"amount"`
		}

		// Parse JSON body
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Create domain transaction
		tx := models.Transaction{
			ID:             uuid.New().String(),
			IdempotencyKey: idempotencyKey,
			FromAccount:    req.FromAccount,
			ToAccount:      req.ToAccount,
			Amount:         req.Amount,
			CreatedAt:      time.Now(),
		}

		// Call domain logic
		exists, err := ledgerService.PostTransaction(context.Background(), tx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if exists {
			http.Error(w, "Duplicate Transaction", http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"Created Transaction"}`))
	})

	http.HandleFunc("/accounts/balance", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		accountId := r.URL.Query().Get("account_id")
		if accountId == "" {
			http.Error(w, "account_id is a mandatory field", http.StatusBadRequest)
			return
		}

		balance, err := ledgerService.GetBalance(accountId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := struct {
			AccountID string          `json:"account_id"`
			Balance   decimal.Decimal `json:"balance"`
		}{
			AccountID: accountId,
			Balance:   balance,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	})

	http.HandleFunc("/ledgerEntries", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		ledgerEntries, err := ledgerService.GetLedgerEntries()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ledgerEntries)

	})
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
