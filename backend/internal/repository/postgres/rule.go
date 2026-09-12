package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"omnishield/internal/rule"
	"omnishield/pkg/errors"
)

type PostgresRuleRepository struct {
	DB *pgxpool.Pool
}

func NewPostgresRuleRepository(db *pgxpool.Pool) *PostgresRuleRepository {
	return &PostgresRuleRepository{DB: db}
}

func (r *PostgresRuleRepository) Create(ctx context.Context, tenantID string, item *rule.AMLRule) error {
	query := `
		INSERT INTO aml_rules (
			tenant_id, rule_code, name, description, rule_type,
			configuration, severity, enabled, version
		) VALUES (
			$1::uuid, $2, $3, $4, $5,
			$6, $7, $8, $9
		)
		RETURNING id::text, created_at, updated_at;
	`

	cfgBytes := item.Configuration
	if len(cfgBytes) == 0 {
		cfgBytes = []byte("{}")
	}

	err := r.DB.QueryRow(
		ctx, query,
		tenantID, item.RuleCode, item.Name, item.Description, item.RuleType,
		cfgBytes, item.Severity, item.Enabled, item.Version,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return errors.Internal(err, "Gagal menambahkan aturan AML baru ke database.")
	}

	return nil
}

func (r *PostgresRuleRepository) GetByID(ctx context.Context, tenantID, id string) (*rule.AMLRule, error) {
	query := `
		SELECT 
			id::text, tenant_id::text, rule_code, name, description,
			rule_type, configuration, severity, enabled, version,
			created_at, updated_at
		FROM aml_rules
		WHERE tenant_id = $1::uuid AND id = $2::uuid
		LIMIT 1;
	`

	var item rule.AMLRule
	err := r.DB.QueryRow(ctx, query, tenantID, id).Scan(
		&item.ID, &item.TenantID, &item.RuleCode, &item.Name, &item.Description,
		&item.RuleType, &item.Configuration, &item.Severity, &item.Enabled, &item.Version,
		&item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound(errors.CodeRuleNotFound, "Aturan AML tidak ditemukan.")
		}
		return nil, errors.Internal(err, "Gagal mengambil data aturan AML.")
	}

	return &item, nil
}

func (r *PostgresRuleRepository) List(ctx context.Context, tenantID string, limit int, cursor string) (*rule.RuleListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var query string
	var rows pgx.Rows
	var err error

	if cursor != "" {
		query = `
			SELECT 
				id::text, tenant_id::text, rule_code, name, description,
				rule_type, configuration, severity, enabled, version,
				created_at, updated_at
			FROM aml_rules
			WHERE tenant_id = $1::uuid AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3;
		`
		rows, err = r.DB.Query(ctx, query, tenantID, cursor, limit+1)
	} else {
		query = `
			SELECT 
				id::text, tenant_id::text, rule_code, name, description,
				rule_type, configuration, severity, enabled, version,
				created_at, updated_at
			FROM aml_rules
			WHERE tenant_id = $1::uuid
			ORDER BY created_at DESC
			LIMIT $2;
		`
		rows, err = r.DB.Query(ctx, query, tenantID, limit+1)
	}

	if err != nil {
		return nil, errors.Internal(err, "Gagal mengambil daftar aturan AML.")
	}
	defer rows.Close()

	list := []rule.AMLRule{}
	for rows.Next() {
		var item rule.AMLRule
		err := rows.Scan(
			&item.ID, &item.TenantID, &item.RuleCode, &item.Name, &item.Description,
			&item.RuleType, &item.Configuration, &item.Severity, &item.Enabled, &item.Version,
			&item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Internal(err, "Gagal membaca baris data aturan AML.")
		}
		list = append(list, item)
	}

	hasMore := false
	nextCursor := ""
	if len(list) > limit {
		hasMore = true
		nextCursor = list[limit-1].CreatedAt.Format(time.RFC3339Nano)
		list = list[:limit]
	}

	return &rule.RuleListResponse{
		Data:       list,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (r *PostgresRuleRepository) Update(ctx context.Context, tenantID, id string, req *rule.UpdateRuleRequest) (*rule.AMLRule, error) {
	cfgBytes := req.Configuration
	if len(cfgBytes) == 0 {
		cfgBytes = []byte("{}")
	}

	query := `
		UPDATE aml_rules
		SET 
			name = $1,
			description = $2,
			configuration = $3,
			severity = $4,
			version = version + 1,
			updated_at = NOW()
		WHERE tenant_id = $5::uuid AND id = $6::uuid
		RETURNING 
			id::text, tenant_id::text, rule_code, name, description,
			rule_type, configuration, severity, enabled, version,
			created_at, updated_at;
	`

	var item rule.AMLRule
	err := r.DB.QueryRow(
		ctx, query,
		req.Name, req.Description, cfgBytes, req.Severity,
		tenantID, id,
	).Scan(
		&item.ID, &item.TenantID, &item.RuleCode, &item.Name, &item.Description,
		&item.RuleType, &item.Configuration, &item.Severity, &item.Enabled, &item.Version,
		&item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound(errors.CodeRuleNotFound, "Aturan AML tidak ditemukan untuk diperbarui.")
		}
		return nil, errors.Internal(err, "Gagal memperbarui aturan AML.")
	}

	return &item, nil
}

func (r *PostgresRuleRepository) ToggleStatus(ctx context.Context, tenantID, id string, enabled bool) (*rule.AMLRule, error) {
	query := `
		UPDATE aml_rules
		SET 
			enabled = $1,
			updated_at = NOW()
		WHERE tenant_id = $2::uuid AND id = $3::uuid
		RETURNING 
			id::text, tenant_id::text, rule_code, name, description,
			rule_type, configuration, severity, enabled, version,
			created_at, updated_at;
	`

	var item rule.AMLRule
	err := r.DB.QueryRow(ctx, query, enabled, tenantID, id).Scan(
		&item.ID, &item.TenantID, &item.RuleCode, &item.Name, &item.Description,
		&item.RuleType, &item.Configuration, &item.Severity, &item.Enabled, &item.Version,
		&item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound(errors.CodeRuleNotFound, "Aturan AML tidak ditemukan untuk diubah statusnya.")
		}
		return nil, errors.Internal(err, "Gagal mengubah status aturan AML.")
	}

	return &item, nil
}
