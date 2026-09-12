package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/alert"
	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/errors"
)

type PostgresAlertRepository struct {
	DB *pgxpool.Pool
}

func NewPostgresAlertRepository(db *pgxpool.Pool) *PostgresAlertRepository {
	return &PostgresAlertRepository{DB: db}
}

func (r *PostgresAlertRepository) GetByID(ctx context.Context, tenantID, id string) (*alert.AlertDetail, error) {
	query := `
		SELECT 
			a.id::text, a.tenant_id::text, a.transaction_id::text, a.risk_assessment_id::text,
			a.risk_score, a.risk_level, a.priority, a.status, a.created_at, a.updated_at,
			t.external_transaction_id, t.account_id, t.counterparty_id, t.amount,
			t.currency, t.transaction_type, t.occurred_at, t.metadata,
			ra.rules_triggered, ra.risk_factors, ra.ai_verdict, ra.ai_explanation
		FROM alerts a
		JOIN transactions t ON a.transaction_id = t.id
		JOIN risk_assessments ra ON a.risk_assessment_id = ra.id
		WHERE a.tenant_id = $1::uuid AND a.id = $2::uuid
		LIMIT 1;
	`

	var d alert.AlertDetail
	err := r.DB.QueryRow(ctx, query, tenantID, id).Scan(
		&d.ID, &d.TenantID, &d.TransactionID, &d.RiskAssessmentID,
		&d.RiskScore, &d.RiskLevel, &d.Priority, &d.Status, &d.CreatedAt, &d.UpdatedAt,
		&d.ExternalTransactionID, &d.AccountID, &d.CounterpartyID, &d.Amount,
		&d.Currency, &d.TransactionType, &d.OccurredAt, &d.TransactionMetadata,
		&d.RulesTriggered, &d.RiskFactors, &d.AIVerdict, &d.AIExplanation,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound(errors.CodeAlertNotFound, "Alert kasus tidak ditemukan.")
		}
		return nil, errors.Internal(err, "Gagal mengambil rincian data alert.")
	}

	return &d, nil
}

func (r *PostgresAlertRepository) List(ctx context.Context, tenantID string, limit int, cursor string) (*alert.AlertListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var query string
	var rows pgx.Rows
	var err error

	if cursor != "" {
		query = `
			SELECT 
				id::text, tenant_id::text, transaction_id::text, risk_assessment_id::text,
				risk_score, risk_level, priority, status, created_at, updated_at
			FROM alerts
			WHERE tenant_id = $1::uuid AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3;
		`
		rows, err = r.DB.Query(ctx, query, tenantID, cursor, limit+1)
	} else {
		query = `
			SELECT 
				id::text, tenant_id::text, transaction_id::text, risk_assessment_id::text,
				risk_score, risk_level, priority, status, created_at, updated_at
			FROM alerts
			WHERE tenant_id = $1::uuid
			ORDER BY created_at DESC
			LIMIT $2;
		`
		rows, err = r.DB.Query(ctx, query, tenantID, limit+1)
	}

	if err != nil {
		return nil, errors.Internal(err, "Gagal mengambil daftar antrean alert.")
	}
	defer rows.Close()

	list := []alert.Alert{}
	for rows.Next() {
		var a alert.Alert
		err := rows.Scan(
			&a.ID, &a.TenantID, &a.TransactionID, &a.RiskAssessmentID,
			&a.RiskScore, &a.RiskLevel, &a.Priority, &a.Status, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Internal(err, "Gagal membaca baris data alert.")
		}
		list = append(list, a)
	}

	hasMore := false
	nextCursor := ""
	if len(list) > limit {
		hasMore = true
		nextCursor = list[limit-1].CreatedAt.Format(time.RFC3339Nano)
		list = list[:limit]
	}

	return &alert.AlertListResponse{
		Data:       list,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
