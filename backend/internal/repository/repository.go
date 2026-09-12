package repository

import (
	"context"
	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/transaction"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *transaction.Transaction) error
}
