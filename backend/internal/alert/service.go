package alert

import (
	"context"

	"omnishield/pkg/errors"
)

type Repository interface {
	GetByID(ctx context.Context, tenantID, id string) (*AlertDetail, error)
	List(ctx context.Context, tenantID string, limit int, cursor string) (*AlertListResponse, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, tenantID, id string) (*AlertDetail, error) {
	if id == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Parameter ID alert wajib disertakan.")
	}
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *Service) List(ctx context.Context, tenantID string, limit int, cursor string) (*AlertListResponse, error) {
	return s.repo.List(ctx, tenantID, limit, cursor)
}
