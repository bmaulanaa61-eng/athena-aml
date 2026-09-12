package webhook

import (
	"encoding/json"
	"time"
)

type WebhookConfig struct {
	ID               string          `json:"id"`
	TenantID         string          `json:"tenant_id"`
	URL              string          `json:"url"`
	SecretEncrypted  string          `json:"secret,omitempty"`
	Enabled          bool            `json:"enabled"`
	SubscribedEvents json.RawMessage `json:"subscribed_events"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type UpsertWebhookConfigRequest struct {
	URL              string   `json:"url"`
	SubscribedEvents []string `json:"subscribed_events,omitempty"`
	Enabled          *bool    `json:"enabled,omitempty"`
}

type WebhookDelivery struct {
	ID           string          `json:"id"`
	TenantID     string          `json:"tenant_id"`
	EventID      string          `json:"event_id"`
	EventType    string          `json:"event_type"`
	Payload      json.RawMessage `json:"payload"`
	Status       string          `json:"status"`
	Attempt      int             `json:"attempt"`
	ResponseCode *int            `json:"response_code,omitempty"`
	LastError    *string         `json:"last_error,omitempty"`
	NextRetryAt  *time.Time      `json:"next_retry_at,omitempty"`
	DeliveredAt  *time.Time      `json:"delivered_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type WebhookDeliveryListResponse struct {
	Data       []WebhookDelivery `json:"data"`
	NextCursor string            `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
}
