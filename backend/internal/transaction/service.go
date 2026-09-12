package transaction

import (
	"context"
	"time"

	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/errors"
)

type Repository interface {
	IngestAtomic(ctx context.Context, tenantID, idempotencyKey string, req *IngestTransactionRequest) (*IngestTransactionResponse, error)
	GetByID(ctx context.Context, tenantID, id string) (*Transaction, error)
	List(ctx context.Context, tenantID string, limit int, cursor string) (*TransactionListResponse, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Ingest(ctx context.Context, tenantID, idempotencyKey string, req *IngestTransactionRequest) (*IngestTransactionResponse, error) {
	if req.ExternalTransactionID == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Field external_transaction_id wajib diisi.")
	}

	if req.AccountID == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Field account_id (rekening/pengirim) wajib diisi.")
	}

	if req.CounterpartyID == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Field counterparty_id (rekening/merchant penerima) wajib diisi.")
	}

	if req.Amount <= 0 {
		return nil, errors.BadRequest(errors.CodeValidationError, "Jumlah amount transaksi harus lebih besar dari 0.")
	}

	if req.Currency == "" {
		req.Currency = "IDR"
	}

	if req.TransactionType == "" {
		req.TransactionType = "PURCHASE"
	}

	if req.OccurredAt.IsZero() {
		req.OccurredAt = time.Now().UTC()
	}

	return s.repo.IngestAtomic(ctx, tenantID, idempotencyKey, req)
}

func (s *Service) GetByID(ctx context.Context, tenantID, id string) (*Transaction, error) {
	if id == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Parameter ID transaksi wajib disertakan.")
	}
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *Service) List(ctx context.Context, tenantID string, limit int, cursor string) (*TransactionListResponse, error) {
	return s.repo.List(ctx, tenantID, limit, cursor)
}
