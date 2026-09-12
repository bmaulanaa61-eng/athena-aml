package transaction

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"omnishield/internal/tenant"
	"omnishield/pkg/errors"
	"omnishield/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) IngestHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.GetTenantID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")

	var req IngestTransactionRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		response.Error(w, r, errors.BadRequest(errors.CodeValidationError, "Format payload JSON transaksi tidak valid."))
		return
	}

	res, err := h.service.Ingest(r.Context(), tenantID, idempotencyKey, &req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Accepted(w, res)
}

func (h *Handler) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.GetTenantID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, r, errors.BadRequest(errors.CodeValidationError, "Parameter id wajib diisi."))
		return
	}

	res, err := h.service.GetByID(r.Context(), tenantID, id)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.OK(w, res)
}

func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.GetTenantID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if parsedLimit, parseErr := strconv.Atoi(limitStr); parseErr == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	cursor := r.URL.Query().Get("cursor")

	res, err := h.service.List(r.Context(), tenantID, limit, cursor)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.OK(w, res)
}
