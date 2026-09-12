package transaction

import (
	"context"
	"testing"
	"time"
)

type fakeRepo struct {
	lastReq *IngestTransactionRequest
}

func (f *fakeRepo) IngestAtomic(ctx context.Context, tenantID, idempotencyKey string, req *IngestTransactionRequest) (*IngestTransactionResponse, error) {
	f.lastReq = req
	return &IngestTransactionResponse{TransactionID: "tx-1", Status: "PROCESSING"}, nil
}

func (f *fakeRepo) GetByID(ctx context.Context, tenantID, id string) (*Transaction, error) {
	return nil, nil
}
func (f *fakeRepo) List(ctx context.Context, tenantID string, limit int, cursor string) (*TransactionListResponse, error) {
	return nil, nil
}

func TestIngestDefaultsCurrencyTypeAndTime(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	res, err := svc.Ingest(context.Background(), "tenant-1", "idem-1", &IngestTransactionRequest{
		ExternalTransactionID: "ext-1",
		AccountID:             "acct-1",
		CounterpartyID:        "merchant-1",
		Amount:                100000,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.TransactionID != "tx-1" {
		t.Fatalf("unexpected response: %+v", res)
	}
	if repo.lastReq.Currency != "IDR" {
		t.Fatalf("expected default currency IDR, got %q", repo.lastReq.Currency)
	}
	if repo.lastReq.TransactionType != "PURCHASE" {
		t.Fatalf("expected default transaction type PURCHASE, got %q", repo.lastReq.TransactionType)
	}
	if repo.lastReq.OccurredAt.IsZero() || time.Since(repo.lastReq.OccurredAt) > time.Minute {
		t.Fatalf("expected occurred_at to be defaulted to now, got %v", repo.lastReq.OccurredAt)
	}
}

func TestIngestRejectsInvalidAmount(t *testing.T) {
	svc := NewService(&fakeRepo{})

	_, err := svc.Ingest(context.Background(), "tenant-1", "idem-1", &IngestTransactionRequest{
		ExternalTransactionID: "ext-1",
		AccountID:             "acct-1",
		CounterpartyID:        "merchant-1",
		Amount:                0,
	})

	if err == nil {
		t.Fatal("expected validation error")
	}
}
