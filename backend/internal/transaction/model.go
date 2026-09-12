package transaction

import (
	"encoding/json"
	"time"
)

type Transaction struct {
	ID                    string          `json:"id"`
	TenantID              string          `json:"tenant_id"`
	ExternalTransactionID string          `json:"external_transaction_id"`
	AccountID             string          `json:"account_id"`
	CounterpartyID        string          `json:"counterparty_id"`
	Amount                float64         `json:"amount"`
	Currency              string          `json:"currency"`
	TransactionType       string          `json:"transaction_type"`
	OccurredAt            time.Time       `json:"occurred_at"`
	Metadata              json.RawMessage `json:"metadata"`
	Status                string          `json:"status"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

type IngestTransactionRequest struct {
	ExternalTransactionID string          `json:"external_transaction_id"`
	AccountID             string          `json:"account_id"`
	CounterpartyID        string          `json:"counterparty_id"`
	Amount                float64         `json:"amount"`
	Currency              string          `json:"currency"`
	TransactionType       string          `json:"transaction_type"`
	OccurredAt            time.Time       `json:"occurred_at"`
	Metadata              json.RawMessage `json:"metadata"`
}

type IngestTransactionResponse struct {
	TransactionID string `json:"transaction_id"`
	Status        string `json:"status"`
	RequestID     string `json:"request_id"`
}

type TransactionListResponse struct {
	Data       []Transaction `json:"data"`
	NextCursor string        `json:"next_cursor,omitempty"`
	HasMore    bool          `json:"has_more"`
}
