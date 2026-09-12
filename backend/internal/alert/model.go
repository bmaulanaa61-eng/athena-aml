package alert

import (
	"encoding/json"
	"time"
)

type Alert struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	TransactionID    string    `json:"transaction_id"`
	RiskAssessmentID string    `json:"risk_assessment_id"`
	RiskScore        int       `json:"risk_score"`
	RiskLevel        string    `json:"risk_level"`
	Priority         string    `json:"priority"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AlertDetail struct {
	Alert
	ExternalTransactionID string          `json:"external_transaction_id"`
	AccountID             string          `json:"account_id"`
	CounterpartyID        string          `json:"counterparty_id"`
	Amount                float64         `json:"amount"`
	Currency              string          `json:"currency"`
	TransactionType       string          `json:"transaction_type"`
	OccurredAt            time.Time       `json:"occurred_at"`
	TransactionMetadata   json.RawMessage `json:"transaction_metadata"`
	RulesTriggered        json.RawMessage `json:"rules_triggered"`
	RiskFactors           json.RawMessage `json:"risk_factors"`
	AIVerdict             string          `json:"ai_verdict"`
	AIExplanation         string          `json:"ai_explanation"`
}

type AlertListResponse struct {
	Data       []Alert `json:"data"`
	NextCursor string  `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}
