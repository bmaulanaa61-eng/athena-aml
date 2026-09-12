package rule

import (
	"encoding/json"
	"time"
)

type AMLRule struct {
	ID            string          `json:"id"`
	TenantID      string          `json:"tenant_id"`
	RuleCode      string          `json:"rule_code"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	RuleType      string          `json:"rule_type"`
	Configuration json.RawMessage `json:"configuration"`
	Severity      string          `json:"severity"`
	Enabled       bool            `json:"enabled"`
	Version       int             `json:"version"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type CreateRuleRequest struct {
	RuleCode      string          `json:"rule_code"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	RuleType      string          `json:"rule_type"`
	Configuration json.RawMessage `json:"configuration"`
	Severity      string          `json:"severity"`
}

type UpdateRuleRequest struct {
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	Configuration json.RawMessage `json:"configuration"`
	Severity      string          `json:"severity"`
}

type ToggleRuleStatusRequest struct {
	Enabled bool `json:"enabled"`
}

type RuleListResponse struct {
	Data       []AMLRule `json:"data"`
	NextCursor string    `json:"next_cursor,omitempty"`
	HasMore    bool      `json:"has_more"`
}
