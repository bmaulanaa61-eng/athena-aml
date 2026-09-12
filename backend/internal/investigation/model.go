package investigation

import (
	"time"
)

type InvestigationNote struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	AlertID   string    `json:"alert_id"`
	UserID    *string   `json:"user_id,omitempty"`
	Author    string    `json:"author"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

type AddNoteRequest struct {
	Author string `json:"author"`
	Note   string `json:"note"`
}

type InvestigationResolution struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	AlertID        string    `json:"alert_id"`
	UserID         *string   `json:"user_id,omitempty"`
	Resolver       string    `json:"resolver"`
	ResolutionType string    `json:"resolution_type"`
	Rationale      string    `json:"rationale"`
	CreatedAt      time.Time `json:"created_at"`
}

type ResolveAlertRequest struct {
	Resolver       string `json:"resolver"`
	ResolutionType string `json:"resolution_type"`
	Rationale      string `json:"rationale"`
}

type CaseTimelineItem struct {
	EventType string    `json:"event_type"`
	Actor     string    `json:"actor"`
	Detail    string    `json:"detail"`
	Timestamp time.Time `json:"timestamp"`
}

type CaseTimelineResponse struct {
	AlertID  string             `json:"alert_id"`
	Status   string             `json:"status"`
	Timeline []CaseTimelineItem `json:"timeline"`
}
