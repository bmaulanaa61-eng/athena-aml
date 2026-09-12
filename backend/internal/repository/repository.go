package repository

import (
	"context"
	"omnishield/internal/transaction"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *transaction.Transaction) error
}
