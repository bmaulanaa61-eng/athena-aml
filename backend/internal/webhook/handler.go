package webhook

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bmaulanaa61-eng/athena-aml/backend/internal/tenant"
	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/errors"
	"github.com/bmaulanaa61-eng/athena-aml/backend/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) UpsertHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.GetTenantID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	var req UpsertWebhookConfigRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		response.Error(w, r, errors.BadRequest(errors.CodeValidationError, "Format payload JSON konfigurasi webhook tidak valid."))
		return
	}

	res, err := h.service.UpsertConfig(r.Context(), tenantID, &req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.OK(w, res)
}

func (h *Handler) GetConfigHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.GetTenantID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	res, err := h.service.GetConfig(r.Context(), tenantID)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.OK(w, res)
}

func (h *Handler) TestPingHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.GetTenantID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	res, err := h.service.TestPing(r.Context(), tenantID)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.OK(w, res)
}

func (h *Handler) DeliveriesHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.GetTenantID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if parsed, parseErr := strconv.Atoi(limitStr); parseErr == nil && parsed > 0 {
			limit = parsed
		}
	}

	cursor := r.URL.Query().Get("cursor")

	res, err := h.service.ListDeliveries(r.Context(), tenantID, limit, cursor)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.OK(w, res)
}
