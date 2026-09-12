package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/middleware"
	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/transaction"
	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/errors"
)

type PostgresTransactionRepository struct {
	DB *pgxpool.Pool
}

func NewPostgresTransactionRepository(db *pgxpool.Pool) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{DB: db}
}

func (r *PostgresTransactionRepository) IngestAtomic(
	ctx context.Context,
	tenantID string,
	idempotencyKey string,
	req *transaction.IngestTransactionRequest,
) (*transaction.IngestTransactionResponse, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, errors.BadRequest(errors.CodeValidationError, "Gagal memproses serialisasi payload transaksi.")
	}
	hasher := sha256.New()
	hasher.Write(reqBytes)
	reqHash := hex.EncodeToString(hasher.Sum(nil))

	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return nil, errors.Internal(err, "Gagal memulai transaksi database.")
	}
	defer tx.Rollback(ctx)

	if idempotencyKey != "" {
		checkIdemQuery := `
			SELECT request_hash, response_snapshot
			FROM idempotency_keys
			WHERE tenant_id = $1::uuid AND idempotency_key = $2 AND expires_at > NOW();
		`
		var existingHash string
		var respSnapshot []byte

		err := tx.QueryRow(ctx, checkIdemQuery, tenantID, idempotencyKey).Scan(&existingHash, &respSnapshot)
		if err == nil {
			if existingHash != reqHash {
				return nil, errors.Conflict(errors.CodeIdempotencyConflict, "Idempotency-Key telah digunakan dengan data transaksi yang berbeda.")
			}

			var cachedResp transaction.IngestTransactionResponse
			if unmarshalErr := json.Unmarshal(respSnapshot, &cachedResp); unmarshalErr == nil {
				return &cachedResp, nil
			}
		} else if err != pgx.ErrNoRows {
			return nil, errors.Internal(err, "Gagal memeriksa status idempotency.")
		}
	}

	insertTxQuery := `
		INSERT INTO transactions (
			tenant_id, external_transaction_id, account_id, counterparty_id,
			amount, currency, transaction_type, occurred_at, metadata, status
		) VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, 'PROCESSING')
		RETURNING id::text, status;
	`

	metaBytes := req.Metadata
	if len(metaBytes) == 0 {
		metaBytes = []byte("{}")
	}

	var generatedTxID, txStatus string
	err = tx.QueryRow(
		ctx, insertTxQuery,
		tenantID, req.ExternalTransactionID, req.AccountID, req.CounterpartyID,
		req.Amount, req.Currency, req.TransactionType, req.OccurredAt, metaBytes,
	).Scan(&generatedTxID, &txStatus)

	if err != nil {
		return nil, errors.Internal(err, "Gagal menyimpan data transaksi ke database.")
	}

	insertJobQuery := `
		INSERT INTO jobs (
			tenant_id, job_type, payload, status
		) VALUES ($1::uuid, 'AML_ASSESSMENT', jsonb_build_object('transaction_id', $2::uuid), 'PENDING');
	`
	_, err = tx.Exec(ctx, insertJobQuery, tenantID, generatedTxID)
	if err != nil {
		return nil, errors.Internal(err, "Gagal mendaftarkan antrean evaluasi AML.")
	}

	requestID := ""
	if reqIDVal, ok := ctx.Value(middleware.RequestIDCtxKey).(string); ok {
		requestID = reqIDVal
	}

	responseObj := transaction.IngestTransactionResponse{
		TransactionID: generatedTxID,
		Status:        txStatus,
		RequestID:     requestID,
	}

	if idempotencyKey != "" {
		snapshotBytes, _ := json.Marshal(responseObj)
		insertIdemQuery := `
			INSERT INTO idempotency_keys (
				tenant_id, idempotency_key, request_hash, transaction_id, response_snapshot, expires_at
			) VALUES ($1::uuid, $2, $3, $4::uuid, $5, NOW() + interval '24 hours');
		`
		_, err = tx.Exec(ctx, insertIdemQuery, tenantID, idempotencyKey, reqHash, generatedTxID, snapshotBytes)
		if err != nil {
			return nil, errors.Internal(err, "Gagal menyimpan kunci idempotency.")
		}
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		return nil, errors.Internal(commitErr, "Gagal melakukan commit transaksi.")
	}

	return &responseObj, nil
}

func (r *PostgresTransactionRepository) GetByID(ctx context.Context, tenantID, id string) (*transaction.Transaction, error) {
	query := `
		SELECT 
			id::text, tenant_id::text, external_transaction_id, account_id, counterparty_id,
			amount, currency, transaction_type, occurred_at, metadata, status, created_at, updated_at
		FROM transactions
		WHERE tenant_id = $1::uuid AND id = $2::uuid
		LIMIT 1;
	`

	var t transaction.Transaction
	err := r.DB.QueryRow(ctx, query, tenantID, id).Scan(
		&t.ID, &t.TenantID, &t.ExternalTransactionID, &t.AccountID, &t.CounterpartyID,
		&t.Amount, &t.Currency, &t.TransactionType, &t.OccurredAt, &t.Metadata,
		&t.Status, &t.CreatedAt, &t.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound(errors.CodeTransactionNotFound, "Transaksi tidak ditemukan.")
		}
		return nil, errors.Internal(err, "Gagal mengambil data transaksi.")
	}

	return &t, nil
}

func (r *PostgresTransactionRepository) List(ctx context.Context, tenantID string, limit int, cursor string) (*transaction.TransactionListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var query string
	var rows pgx.Rows
	var err error

	if cursor != "" {
		query = `
			SELECT 
				id::text, tenant_id::text, external_transaction_id, account_id, counterparty_id,
				amount, currency, transaction_type, occurred_at, metadata, status, created_at, updated_at
			FROM transactions
			WHERE tenant_id = $1::uuid AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3;
		`
		rows, err = r.DB.Query(ctx, query, tenantID, cursor, limit+1)
	} else {
		query = `
			SELECT 
				id::text, tenant_id::text, external_transaction_id, account_id, counterparty_id,
				amount, currency, transaction_type, occurred_at, metadata, status, created_at, updated_at
			FROM transactions
			WHERE tenant_id = $1::uuid
			ORDER BY created_at DESC
			LIMIT $2;
		`
		rows, err = r.DB.Query(ctx, query, tenantID, limit+1)
	}

	if err != nil {
		return nil, errors.Internal(err, "Gagal mengambil daftar transaksi.")
	}
	defer rows.Close()

	list := []transaction.Transaction{}
	for rows.Next() {
		var t transaction.Transaction
		err := rows.Scan(
			&t.ID, &t.TenantID, &t.ExternalTransactionID, &t.AccountID, &t.CounterpartyID,
			&t.Amount, &t.Currency, &t.TransactionType, &t.OccurredAt, &t.Metadata,
			&t.Status, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, errors.Internal(err, "Gagal memproses baris transaksi.")
		}
		list = append(list, t)
	}

	hasMore := false
	nextCursor := ""
	if len(list) > limit {
		hasMore = true
		nextCursor = list[limit-1].CreatedAt.Format(time.RFC3339Nano)
		list = list[:limit]
	}

	return &transaction.TransactionListResponse{
		Data:       list,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
