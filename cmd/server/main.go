package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	interfaces "github.com/sheikh-saqib/distributed-payments-ledger-system/internal/interfaces"
	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/ledger"
	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/models"
	"github.com/sheikh-saqib/distributed-payments-ledger-system/internal/storage/memory"
	"github.com/shopspring/decimal"
)

func main() {

	var store interfaces.LedgerStore = memory.NewMemoryLedgerStore()
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
		err := ledgerService.PostTransaction(context.Background(), tx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// if transaction.Replayed {
		// 	w.WriteHeader(http.StatusOK)
		// } else {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"Created Transaction"}`))
		// }

		// if err := json.NewEncoder(w).Encode(transaction); err != nil {
		// 	http.Error(w, "failed to encode response", http.StatusInternalServerError)
		// 	return
		// }
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
