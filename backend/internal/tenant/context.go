package tenant

import (
	"context"

	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/errors"
)

type contextKey string

const (
	tenantIDKey contextKey = "tenant_id"
)

func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

func GetTenantID(ctx context.Context) (string, error) {
	val := ctx.Value(tenantIDKey)
	if val == nil {
		return "", errors.Unauthorized("Konteks tenant tidak ditemukan pada request.")
	}

	tenantID, ok := val.(string)
	if !ok || tenantID == "" {
		return "", errors.Unauthorized("Format identitas tenant pada konteks tidak valid.")
	}

	return tenantID, nil
}
