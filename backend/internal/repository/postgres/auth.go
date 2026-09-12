package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"omnishield/internal/auth"
	"omnishield/pkg/errors"
)

type PostgresAuthRepository struct {
	DB *pgxpool.Pool
}

func NewPostgresAuthRepository(db *pgxpool.Pool) *PostgresAuthRepository {
	return &PostgresAuthRepository{DB: db}
}

func (r *PostgresAuthRepository) CreateUser(ctx context.Context, email, passwordHash, role string) (*auth.User, error) {
	var tenantID string
	err := r.DB.QueryRow(ctx, `SELECT id::text FROM tenants LIMIT 1;`).Scan(&tenantID)
	if err != nil {
		return nil, errors.Internal(err, "Tenant organisasi belum tersedia.")
	}

	query := `
		INSERT INTO users (
			tenant_id, email, password_hash, role, status
		) VALUES (
			$1::uuid, $2, $3, $4, 'ACTIVE'
		)
		ON CONFLICT (tenant_id, email) DO UPDATE
		SET password_hash = EXCLUDED.password_hash, role = EXCLUDED.role, status = 'ACTIVE', updated_at = NOW()
		RETURNING id::text, tenant_id::text, email, password_hash, role, status, last_login_at, created_at, updated_at;
	`

	var u auth.User
	err = r.DB.QueryRow(ctx, query, tenantID, email, passwordHash, role).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Internal(err, "Gagal membuat akun user analis.")
	}

	return &u, nil
}

func (r *PostgresAuthRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
		SELECT id::text, tenant_id::text, email, password_hash, role, status, last_login_at, created_at, updated_at
		FROM users
		WHERE LOWER(email) = LOWER($1)
		LIMIT 1;
	`

	var u auth.User
	err := r.DB.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFound(errors.CodeResourceNotFound, "User tidak ditemukan.")
		}
		return nil, errors.Internal(err, "Gagal mengambil data user.")
	}

	return &u, nil
}

func (r *PostgresAuthRepository) CreateSession(ctx context.Context, tenantID, userID, sessionHash string, expiresAt time.Time) (*auth.Session, error) {
	query := `
		INSERT INTO sessions (
			tenant_id, user_id, session_hash, expires_at
		) VALUES (
			$1::uuid, $2::uuid, $3, $4
		)
		RETURNING id::text, tenant_id::text, user_id::text, session_hash, expires_at, last_seen_at, created_at;
	`

	var s auth.Session
	err := r.DB.QueryRow(ctx, query, tenantID, userID, sessionHash, expiresAt).Scan(
		&s.ID, &s.TenantID, &s.UserID, &s.SessionHash, &s.ExpiresAt, &s.LastSeenAt, &s.CreatedAt,
	)
	if err != nil {
		return nil, errors.Internal(err, "Gagal menyimpan sesi login.")
	}

	return &s, nil
}

func (r *PostgresAuthRepository) GetSessionUser(ctx context.Context, sessionHash string) (*auth.User, error) {
	query := `
		SELECT 
			u.id::text, u.tenant_id::text, u.email, u.password_hash, u.role, u.status, u.last_login_at, u.created_at, u.updated_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.session_hash = $1 AND s.revoked_at IS NULL AND s.expires_at > NOW()
		LIMIT 1;
	`

	var u auth.User
	err := r.DB.QueryRow(ctx, query, sessionHash).Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.Unauthorized("Sesi login tidak valid atau sudah kedaluwarsa.")
		}
		return nil, errors.Internal(err, "Gagal memverifikasi sesi login.")
	}

	_, _ = r.DB.Exec(ctx, `UPDATE sessions SET last_seen_at = NOW() WHERE session_hash = $1;`, sessionHash)

	return &u, nil
}

func (r *PostgresAuthRepository) RevokeSession(ctx context.Context, sessionHash string) error {
	query := `UPDATE sessions SET revoked_at = NOW() WHERE session_hash = $1 AND revoked_at IS NULL;`
	_, err := r.DB.Exec(ctx, query, sessionHash)
	if err != nil {
		return errors.Internal(err, "Gagal mencabut sesi login.")
	}
	return nil
}

func (r *PostgresAuthRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1::uuid;`
	_, err := r.DB.Exec(ctx, query, userID)
	return err
}
