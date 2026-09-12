package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/config"
	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/database"
	"github.com/bmaulanaa61-eng/athena-aml/backend/routes"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	router := routes.SetupRoutes(db)

	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Server berjalan di http://localhost:" + cfg.AppPort)
		if srvErr := server.ListenAndServe(); srvErr != nil && srvErr != http.ErrServerClosed {
			log.Fatalf("Server mati mendadak: %v", srvErr)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Mematikan server secara graceful...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if shutdownErr := server.Shutdown(ctx); shutdownErr != nil {
		log.Fatalf("Gagal graceful shutdown: %v", shutdownErr)
	}

	log.Println("Server berhasil dihentikan dengan aman ✅")
}
