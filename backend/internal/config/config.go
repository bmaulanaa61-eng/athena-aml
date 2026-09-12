package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	AppPort    string
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: Tidak menemukan file .env, menggunakan OS environment bawaan.")
	}

	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),
		AppPort:    os.Getenv("APP_PORT"),
	}

	if cfg.DBSSLMode == "" {
		cfg.DBSSLMode = "disable"
	}
	if cfg.AppPort == "" {
		cfg.AppPort = "8080"
	}

	missing := []string{}
	if cfg.DBHost == "" {
		missing = append(missing, "DB_HOST")
	}
	if cfg.DBPort == "" {
		missing = append(missing, "DB_PORT")
	}
	if cfg.DBUser == "" {
		missing = append(missing, "DB_USER")
	}
	if cfg.DBPassword == "" {
		missing = append(missing, "DB_PASSWORD")
	}
	if cfg.DBName == "" {
		missing = append(missing, "DB_NAME")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required env vars: %v", missing)
	}
	return cfg, nil
}
