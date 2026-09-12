package auth

import (
	"context"
	"testing"
	"time"
)

type fakeAuthRepo struct {
	createdHash string
	user        *User
}

func (f *fakeAuthRepo) CreateUser(ctx context.Context, email, passwordHash, role string) (*User, error) {
	f.createdHash = passwordHash
	return &User{ID: "u1", TenantID: "t1", Email: email, PasswordHash: passwordHash, Role: role, Status: "ACTIVE"}, nil
}
func (f *fakeAuthRepo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return f.user, nil
}
func (f *fakeAuthRepo) CreateSession(ctx context.Context, tenantID, userID, sessionHash string, expiresAt time.Time) (*Session, error) {
	return &Session{ID: "s1", TenantID: tenantID, UserID: userID, SessionHash: sessionHash, ExpiresAt: expiresAt}, nil
}
func (f *fakeAuthRepo) GetSessionUser(ctx context.Context, sessionHash string) (*User, error) {
	return f.user, nil
}
func (f *fakeAuthRepo) RevokeSession(ctx context.Context, sessionHash string) error { return nil }
func (f *fakeAuthRepo) UpdateLastLogin(ctx context.Context, userID string) error    { return nil }

func TestRegisterHashesPasswordAndDefaultsRole(t *testing.T) {
	repo := &fakeAuthRepo{}
	svc := NewService(repo)

	user, err := svc.Register(context.Background(), &RegisterRequest{Email: "admin@example.com", Password: "Password123!"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Role != "ADMIN" {
		t.Fatalf("expected ADMIN role, got %s", user.Role)
	}
	if repo.createdHash == "Password123!" || repo.createdHash == "" {
		t.Fatalf("password was not hashed")
	}
}

func TestRegisterRejectsWeakPassword(t *testing.T) {
	svc := NewService(&fakeAuthRepo{})
	_, err := svc.Register(context.Background(), &RegisterRequest{Email: "admin@example.com", Password: "short"})
	if err == nil {
		t.Fatal("expected validation error for weak password")
	}
}
