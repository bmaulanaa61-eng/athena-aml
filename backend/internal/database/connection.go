package database

import (
	"context"
	"fmt"
	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDB(cfg *config.Config) (*pgxpool.Pool, error) {

	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("gagal Membuat connection pool: %w", err)
	}

	pingErr := pool.Ping(context.Background())
	if pingErr != nil {
		return nil, fmt.Errorf("gagal ping database: %w", pingErr)
	}

	fmt.Printf("Koneksi database berhasil ✅")
	return pool, nil
}
