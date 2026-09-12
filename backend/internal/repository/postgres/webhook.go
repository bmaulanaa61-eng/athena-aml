package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/webhook"
	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/errors"
)

type PostgresWebhookRepository struct {
	DB *pgxpool.Pool
}

func NewPostgresWebhookRepository(db *pgxpool.Pool) *PostgresWebhookRepository {
	return &PostgresWebhookRepository{DB: db}
}

func (r *PostgresWebhookRepository) UpsertConfig(ctx context.Context, tenantID string, req *webhook.UpsertWebhookConfigRequest, secret string) (*webhook.WebhookConfig, error) {
	eventsJSON, _ := json.Marshal(req.SubscribedEvents)
	if len(req.SubscribedEvents) == 0 {
		eventsJSON = []byte(`["AML_ALERT_CREATED", "AML_ALERT_RESOLVED"]`)
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	var existingID string
	checkQuery := `SELECT id::text FROM webhook_configs WHERE tenant_id = $1::uuid LIMIT 1;`
	err := r.DB.QueryRow(ctx, checkQuery, tenantID).Scan(&existingID)

	var cfg webhook.WebhookConfig
	if err == pgx.ErrNoRows {
		insertQuery := `
			INSERT INTO webhook_configs (
				tenant_id, url, secret_encrypted, enabled, subscribed_events
			) VALUES (
				$1::uuid, $2, $3, $4, $5::jsonb
			)
			RETURNING id::text, tenant_id::text, url, secret_encrypted, enabled, subscribed_events, created_at, updated_at;
		`
		err = r.DB.QueryRow(ctx, insertQuery, tenantID, req.URL, secret, enabled, eventsJSON).Scan(
			&cfg.ID, &cfg.TenantID, &cfg.URL, &cfg.SecretEncrypted, &cfg.Enabled, &cfg.SubscribedEvents, &cfg.CreatedAt, &cfg.UpdatedAt,
		)
	} else if err == nil {
		updateQuery := `
			UPDATE webhook_configs
			SET url = $1, enabled = $2, subscribed_events = $3::jsonb, updated_at = NOW()
			WHERE id = $4::uuid
			RETURNING id::text, tenant_id::text, url, secret_encrypted, enabled, subscribed_events, created_at, updated_at;
		`
		err = r.DB.QueryRow(ctx, updateQuery, req.URL, enabled, eventsJSON, existingID).Scan(
			&cfg.ID, &cfg.TenantID, &cfg.URL, &cfg.SecretEncrypted, &cfg.Enabled, &cfg.SubscribedEvents, &cfg.CreatedAt, &cfg.UpdatedAt,
		)
	}

	if err != nil {
		return nil, errors.Internal(err, "Gagal menyimpan konfigurasi webhook.")
	}

	return &cfg, nil
}

func (r *PostgresWebhookRepository) GetConfig(ctx context.Context, tenantID string) (*webhook.WebhookConfig, error) {
	query := `
		SELECT id::text, tenant_id::text, url, secret_encrypted, enabled, subscribed_events, created_at, updated_at
		FROM webhook_configs
		WHERE tenant_id = $1::uuid
		LIMIT 1;
	`

	var cfg webhook.WebhookConfig
	err := r.DB.QueryRow(ctx, query, tenantID).Scan(
		&cfg.ID, &cfg.TenantID, &cfg.URL, &cfg.SecretEncrypted, &cfg.Enabled, &cfg.SubscribedEvents, &cfg.CreatedAt, &cfg.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound(errors.CodeResourceNotFound, "Konfigurasi webhook belum didaftarkan.")
		}
		return nil, errors.Internal(err, "Gagal mengambil konfigurasi webhook.")
	}

	return &cfg, nil
}

func (r *PostgresWebhookRepository) CreateDelivery(ctx context.Context, tenantID, eventID, eventType string, payload []byte, status string, responseCode int) (*webhook.WebhookDelivery, error) {
	now := time.Now().UTC()
	query := `
		INSERT INTO webhook_deliveries (
			tenant_id, event_id, event_type, payload, status, attempt, response_code, delivered_at
		) VALUES (
			$1::uuid, $2, $3, $4::jsonb, $5, 1, $6, $7
		)
		RETURNING id::text, tenant_id::text, event_id, event_type, payload, status, attempt, response_code, last_error, next_retry_at, delivered_at, created_at, updated_at;
	`

	var d webhook.WebhookDelivery
	err := r.DB.QueryRow(ctx, query, tenantID, eventID, eventType, payload, status, responseCode, now).Scan(
		&d.ID, &d.TenantID, &d.EventID, &d.EventType, &d.Payload, &d.Status, &d.Attempt, &d.ResponseCode, &d.LastError, &d.NextRetryAt, &d.DeliveredAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Internal(err, "Gagal mencatat log pengiriman webhook.")
	}

	return &d, nil
}

func (r *PostgresWebhookRepository) ListDeliveries(ctx context.Context, tenantID string, limit int, cursor string) (*webhook.WebhookDeliveryListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var query string
	var rows pgx.Rows
	var err error

	if cursor != "" {
		query = `
			SELECT id::text, tenant_id::text, event_id, event_type, payload, status, attempt, response_code, last_error, next_retry_at, delivered_at, created_at, updated_at
			FROM webhook_deliveries
			WHERE tenant_id = $1::uuid AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3;
		`
		rows, err = r.DB.Query(ctx, query, tenantID, cursor, limit+1)
	} else {
		query = `
			SELECT id::text, tenant_id::text, event_id, event_type, payload, status, attempt, response_code, last_error, next_retry_at, delivered_at, created_at, updated_at
			FROM webhook_deliveries
			WHERE tenant_id = $1::uuid
			ORDER BY created_at DESC
			LIMIT $2;
		`
		rows, err = r.DB.Query(ctx, query, tenantID, limit+1)
	}

	if err != nil {
		return nil, errors.Internal(err, "Gagal mengambil riwayat pengiriman webhook.")
	}
	defer rows.Close()

	list := []webhook.WebhookDelivery{}
	for rows.Next() {
		var d webhook.WebhookDelivery
		if err := rows.Scan(
			&d.ID, &d.TenantID, &d.EventID, &d.EventType, &d.Payload, &d.Status, &d.Attempt, &d.ResponseCode, &d.LastError, &d.NextRetryAt, &d.DeliveredAt, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, errors.Internal(err, "Gagal membaca baris pengiriman webhook.")
		}
		list = append(list, d)
	}

	hasMore := false
	nextCursor := ""
	if len(list) > limit {
		hasMore = true
		nextCursor = list[limit-1].CreatedAt.Format(time.RFC3339Nano)
		list = list[:limit]
	}

	return &webhook.WebhookDeliveryListResponse{
		Data:       list,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
