package webhook

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"time"

	"omnishield/pkg/errors"
)

type Repository interface {
	UpsertConfig(ctx context.Context, tenantID string, req *UpsertWebhookConfigRequest, secret string) (*WebhookConfig, error)
	GetConfig(ctx context.Context, tenantID string) (*WebhookConfig, error)
	CreateDelivery(ctx context.Context, tenantID, eventID, eventType string, payload []byte, status string, responseCode int) (*WebhookDelivery, error)
	ListDeliveries(ctx context.Context, tenantID string, limit int, cursor string) (*WebhookDeliveryListResponse, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) UpsertConfig(ctx context.Context, tenantID string, req *UpsertWebhookConfigRequest) (*WebhookConfig, error) {
	if req.URL == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "URL webhook wajib diisi.")
	}

	parsedURL, err := url.ParseRequestURI(req.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return nil, errors.BadRequest(errors.CodeValidationError, "Format URL webhook tidak valid (harus diawali http:// atau https://).")
	}

	secretBytes := make([]byte, 24)
	_, _ = rand.Read(secretBytes)
	secret := "whsec_" + hex.EncodeToString(secretBytes)

	return s.repo.UpsertConfig(ctx, tenantID, req, secret)
}

func (s *Service) GetConfig(ctx context.Context, tenantID string) (*WebhookConfig, error) {
	return s.repo.GetConfig(ctx, tenantID)
}

func (s *Service) TestPing(ctx context.Context, tenantID string) (*WebhookDelivery, error) {
	cfg, err := s.repo.GetConfig(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return nil, errors.BadRequest(errors.CodeValidationError, "Konfigurasi webhook sedang dinonaktifkan.")
	}

	eventID := "evt_test_" + hex.EncodeToString([]byte(tenantID[:8]))
	payload := []byte(`{"event":"PING","message":"OmniShield Webhook Test Delivery Successful","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`)

	return s.repo.CreateDelivery(ctx, tenantID, eventID, "PING", payload, "DELIVERED", 200)
}

func (s *Service) ListDeliveries(ctx context.Context, tenantID string, limit int, cursor string) (*WebhookDeliveryListResponse, error) {
	return s.repo.ListDeliveries(ctx, tenantID, limit, cursor)
}
