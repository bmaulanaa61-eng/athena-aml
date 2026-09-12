package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/alert"
	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/auth"
	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/investigation"
	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/middleware"
	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/repository/postgres"
	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/rule"
	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/transaction"
	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/webhook"
	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/response"
)

func SetupRoutes(db *pgxpool.Pool) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery)
	r.Use(middleware.BodyLimit(1 << 20))

	registerHealthRoutes(r, db)
	registerAuthRoutes(r, db)
	registerTransactionRoutes(r, db)
	registerRuleRoutes(r, db)
	registerAlertRoutes(r, db)
	registerWebhookRoutes(r, db)

	return r
}

func registerHealthRoutes(r chi.Router, db *pgxpool.Pool) {
	r.Get("/v1/health", func(w http.ResponseWriter, req *http.Request) {
		dbStatus := "UP"
		if db == nil {
			dbStatus = "DOWN"
		} else if pingErr := db.Ping(req.Context()); pingErr != nil {
			dbStatus = "DOWN"
		}
		response.OK(w, map[string]string{
			"status":   "OK",
			"database": dbStatus,
			"service":  "athena-api",
		})
	})
}

func registerAuthRoutes(r chi.Router, db *pgxpool.Pool) {
	repo := postgres.NewPostgresAuthRepository(db)
	svc := auth.NewService(repo)
	h := auth.NewHandler(svc)

	r.Route("/v1/auth", func(r chi.Router) {
		r.Post("/register", h.RegisterHandler)
		r.Post("/login", h.LoginHandler)
		r.Get("/me", h.MeHandler)
		r.Post("/logout", h.LogoutHandler)
	})
}

func registerTransactionRoutes(r chi.Router, db *pgxpool.Pool) {
	repo := postgres.NewPostgresTransactionRepository(db)
	svc := transaction.NewService(repo)
	h := transaction.NewHandler(svc)

	r.Route("/v1/transactions", func(r chi.Router) {
		r.Use(auth.RequireAPIKey(db))
		r.Post("/", h.IngestHandler)
		r.Get("/", h.ListHandler)
		r.Get("/{id}", h.GetByIDHandler)
	})
}

func registerRuleRoutes(r chi.Router, db *pgxpool.Pool) {
	repo := postgres.NewPostgresRuleRepository(db)
	svc := rule.NewService(repo)
	h := rule.NewHandler(svc)

	r.Route("/v1/rules", func(r chi.Router) {
		r.Use(auth.RequireAPIKey(db))
		r.Post("/", h.CreateHandler)
		r.Get("/", h.ListHandler)
		r.Get("/{id}", h.GetByIDHandler)
		r.Put("/{id}", h.UpdateHandler)
		r.Patch("/{id}/status", h.ToggleStatusHandler)
	})
}

func registerAlertRoutes(r chi.Router, db *pgxpool.Pool) {
	alertRepo := postgres.NewPostgresAlertRepository(db)
	alertSvc := alert.NewService(alertRepo)
	alertHandler := alert.NewHandler(alertSvc)

	invRepo := postgres.NewPostgresInvestigationRepository(db)
	invSvc := investigation.NewService(invRepo)
	invHandler := investigation.NewHandler(invSvc)

	r.Route("/v1/alerts", func(r chi.Router) {
		r.Use(auth.RequireAPIKey(db))
		r.Get("/", alertHandler.ListHandler)
		r.Get("/{id}", alertHandler.GetByIDHandler)
		r.Post("/{id}/notes", invHandler.AddNoteHandler)
		r.Get("/{id}/notes", invHandler.GetNotesHandler)
		r.Post("/{id}/resolve", invHandler.ResolveHandler)
		r.Get("/{id}/timeline", invHandler.GetTimelineHandler)
	})
}

func registerWebhookRoutes(r chi.Router, db *pgxpool.Pool) {
	repo := postgres.NewPostgresWebhookRepository(db)
	svc := webhook.NewService(repo)
	h := webhook.NewHandler(svc)

	r.Route("/v1/webhooks", func(r chi.Router) {
		r.Use(auth.RequireAPIKey(db))
		r.Post("/", h.UpsertHandler)
		r.Get("/", h.GetConfigHandler)
		r.Post("/test", h.TestPingHandler)
		r.Get("/deliveries", h.DeliveriesHandler)
	})
}
