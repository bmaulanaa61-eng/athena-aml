package investigation

import (
	"context"

	"omnishield/pkg/errors"
)

type Repository interface {
	AddNote(ctx context.Context, tenantID, alertID string, req *AddNoteRequest) (*InvestigationNote, error)
	GetNotes(ctx context.Context, tenantID, alertID string) ([]InvestigationNote, error)
	ResolveCase(ctx context.Context, tenantID, alertID string, req *ResolveAlertRequest) (*InvestigationResolution, error)
	GetTimeline(ctx context.Context, tenantID, alertID string) (*CaseTimelineResponse, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) AddNote(ctx context.Context, tenantID, alertID string, req *AddNoteRequest) (*InvestigationNote, error) {
	if alertID == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Parameter ID alert wajib disertakan.")
	}
	if req.Note == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Isi catatan investigasi (note) tidak boleh kosong.")
	}
	if req.Author == "" {
		req.Author = "Compliance Officer"
	}
	return s.repo.AddNote(ctx, tenantID, alertID, req)
}

func (s *Service) GetNotes(ctx context.Context, tenantID, alertID string) ([]InvestigationNote, error) {
	if alertID == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Parameter ID alert wajib disertakan.")
	}
	return s.repo.GetNotes(ctx, tenantID, alertID)
}

func (s *Service) ResolveCase(ctx context.Context, tenantID, alertID string, req *ResolveAlertRequest) (*InvestigationResolution, error) {
	if alertID == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Parameter ID alert wajib disertakan.")
	}
	if req.ResolutionType != "FALSE_POSITIVE" && req.ResolutionType != "CONFIRMED_FRAUD" && req.ResolutionType != "SUSPICIOUS_REPORTED" {
		return nil, errors.BadRequest(errors.CodeAlertStateInvalid, "Tipe resolusi harus salah satu dari: FALSE_POSITIVE, CONFIRMED_FRAUD, SUSPICIOUS_REPORTED.")
	}
	if req.Rationale == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Alasan keputusan (rationale) wajib diisi untuk keperluan audit hukum.")
	}
	if req.Resolver == "" {
		req.Resolver = "Compliance Lead"
	}
	return s.repo.ResolveCase(ctx, tenantID, alertID, req)
}

func (s *Service) GetTimeline(ctx context.Context, tenantID, alertID string) (*CaseTimelineResponse, error) {
	if alertID == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Parameter ID alert wajib disertakan.")
	}
	return s.repo.GetTimeline(ctx, tenantID, alertID)
}
