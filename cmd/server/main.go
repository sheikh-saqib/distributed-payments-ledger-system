package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

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

		// Request DTO (like a C# request model)
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
			ID:          time.Now().Format("20060102150405"),
			FromAccount: req.FromAccount,
			ToAccount:   req.ToAccount,
			Amount:      req.Amount,
			CreatedAt:   time.Now(),
		}

		// Call domain logic
		if err := ledgerService.PostTransaction(context.Background(), tx); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status":"transaction posted"}`))
	})

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
