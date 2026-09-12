package alert

import (
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

func (h *Handler) GetByIDHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.GetTenantID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		response.Error(w, r, errors.BadRequest(errors.CodeValidationError, "Parameter id alert wajib disertakan."))
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
		if parsed, parseErr := strconv.Atoi(limitStr); parseErr == nil && parsed > 0 {
			limit = parsed
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
