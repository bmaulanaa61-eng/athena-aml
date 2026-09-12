package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/tenant"
	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/errors"
	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/response"
)

func RequireAPIKey(db *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawKey := r.Header.Get("X-API-KEY")
			if rawKey == "" {
				response.Error(w, r, errors.Unauthorized("Header X-API-KEY wajib disertakan"))
				return
			}
			if len(rawKey) < 12 {
				response.Error(w, r, errors.Unauthorized("Format Header X-API-KEY tidak valid"))
				return
			}
			keyPrefix := rawKey[:12]
			query := `
				SELECT 
					a.id::text, 
					a.tenant_id::text, 
					a.key_hash, 
					a.status, 
					t.status, 
					a.expires_at 
				FROM api_keys a 
				JOIN tenants t ON a.tenant_id = t.id
				WHERE a.key_prefix = $1
				LIMIT 1;
			`
			var apiKeyID, tenantID, keyHash, keyStatus, tenantStatus string
			var expiresAt *time.Time
			err := db.QueryRow(r.Context(), query, keyPrefix).Scan(
				&apiKeyID,
				&tenantID,
				&keyHash,
				&keyStatus,
				&tenantStatus,
				&expiresAt,
			)
			if err != nil {
				if err == pgx.ErrNoRows {
					response.Error(w, r, errors.Unauthorized("API Key tidak valid atau tidak terdaftar"))
					return
				}
				response.Error(w, r, errors.Internal(err, "Gagal memverifikasi API key ke database"))
				return
			}

			if bcryptErr := bcrypt.CompareHashAndPassword([]byte(keyHash), []byte(rawKey)); bcryptErr != nil {
				response.Error(w, r, errors.Unauthorized("API Key tidak valid atau tidak terdaftar"))
				return
			}

			if keyStatus != "ACTIVE" {
				response.Error(w, r, errors.Unauthorized("API Key dinonaktifkan atau dicabut"))
				return
			}
			if tenantStatus != "ACTIVE" {
				response.Error(w, r, errors.Forbidden("Organisasi tenant sedang berstatus non-aktif"))
				return
			}
			if expiresAt != nil && expiresAt.Before(time.Now().UTC()) {
				response.Error(w, r, errors.Unauthorized("API Key telah kedaluwarsa"))
				return
			}
			go func(keyID string) {
				updateQuery := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
				_, _ = db.Exec(context.Background(), updateQuery, keyID)
			}(apiKeyID)
			ctx := tenant.WithTenantID(r.Context(), tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
