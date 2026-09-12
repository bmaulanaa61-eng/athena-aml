package investigation

import (
	"encoding/json"
	"net/http"

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

func (h *Handler) AddNoteHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.GetTenantID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	alertID := chi.URLParam(r, "id")

	var req AddNoteRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		response.Error(w, r, errors.BadRequest(errors.CodeValidationError, "Format payload JSON catatan tidak valid."))
		return
	}

	res, err := h.service.AddNote(r.Context(), tenantID, alertID, &req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.Created(w, res)
}

func (h *Handler) GetNotesHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.GetTenantID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	alertID := chi.URLParam(r, "id")
	res, err := h.service.GetNotes(r.Context(), tenantID, alertID)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.OK(w, res)
}

func (h *Handler) ResolveHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.GetTenantID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	alertID := chi.URLParam(r, "id")

	var req ResolveAlertRequest
	if decodeErr := json.NewDecoder(r.Body).Decode(&req); decodeErr != nil {
		response.Error(w, r, errors.BadRequest(errors.CodeValidationError, "Format payload JSON resolusi tidak valid."))
		return
	}

	res, err := h.service.ResolveCase(r.Context(), tenantID, alertID, &req)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.OK(w, res)
}

func (h *Handler) GetTimelineHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, err := tenant.GetTenantID(r.Context())
	if err != nil {
		response.Error(w, r, err)
		return
	}

	alertID := chi.URLParam(r, "id")
	res, err := h.service.GetTimeline(r.Context(), tenantID, alertID)
	if err != nil {
		response.Error(w, r, err)
		return
	}

	response.OK(w, res)
}
