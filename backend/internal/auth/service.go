package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/mail"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/errors"
)

type Repository interface {
	CreateUser(ctx context.Context, email, passwordHash, role string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateSession(ctx context.Context, tenantID, userID, sessionHash string, expiresAt time.Time) (*Session, error)
	GetSessionUser(ctx context.Context, sessionHash string) (*User, error)
	RevokeSession(ctx context.Context, sessionHash string) error
	UpdateLastLogin(ctx context.Context, userID string) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(ctx context.Context, req *RegisterRequest) (*User, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Email dan password wajib diisi.")
	}

	if _, parseErr := mail.ParseAddress(req.Email); parseErr != nil {
		return nil, errors.BadRequest(errors.CodeValidationError, "Format alamat email tidak valid.")
	}

	if len(req.Password) < 8 {
		return nil, errors.BadRequest(errors.CodeValidationError, "Password minimal 8 karakter.")
	}

	if req.Role == "" {
		req.Role = "ADMIN"
	}
	if req.Role != "ADMIN" && req.Role != "COMPLIANCE" && req.Role != "VIEWER" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Role harus salah satu dari: ADMIN, COMPLIANCE, VIEWER.")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, errors.Internal(err, "Gagal mengenkripsi password dengan Bcrypt.")
	}

	return s.repo.CreateUser(ctx, req.Email, string(hash), req.Role)
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, string, time.Time, error) {
	if email == "" || password == "" {
		return nil, "", time.Time{}, errors.BadRequest(errors.CodeValidationError, "Email dan password wajib diisi.")
	}

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", time.Time{}, errors.Unauthorized("Email atau password salah.")
	}

	if user.Status != "ACTIVE" {
		return nil, "", time.Time{}, errors.Unauthorized("Akun Anda sedang dinonaktifkan.")
	}

	if bcryptErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); bcryptErr != nil {
		return nil, "", time.Time{}, errors.Unauthorized("Email atau password salah.")
	}

	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	rawToken := "os_sess_" + hex.EncodeToString(tokenBytes)

	hash := sha256.Sum256([]byte(rawToken))
	sessionHash := hex.EncodeToString(hash[:])

	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)

	_, err = s.repo.CreateSession(ctx, user.TenantID, user.ID, sessionHash, expiresAt)
	if err != nil {
		return nil, "", time.Time{}, errors.Internal(err, "Gagal membuat sesi login.")
	}

	_ = s.repo.UpdateLastLogin(ctx, user.ID)

	return user, rawToken, expiresAt, nil
}

func (s *Service) GetMe(ctx context.Context, rawToken string) (*User, error) {
	if rawToken == "" {
		return nil, errors.Unauthorized("Sesi tidak ditemukan atau telah berakhir.")
	}

	hash := sha256.Sum256([]byte(rawToken))
	sessionHash := hex.EncodeToString(hash[:])

	return s.repo.GetSessionUser(ctx, sessionHash)
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}

	hash := sha256.Sum256([]byte(rawToken))
	sessionHash := hex.EncodeToString(hash[:])

	return s.repo.RevokeSession(ctx, sessionHash)
}
