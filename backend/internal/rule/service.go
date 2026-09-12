package rule

import (
	"context"

	"omnishield/pkg/errors"
)

type Repository interface {
	Create(ctx context.Context, tenantID string, r *AMLRule) error
	GetByID(ctx context.Context, tenantID, id string) (*AMLRule, error)
	List(ctx context.Context, tenantID string, limit int, cursor string) (*RuleListResponse, error)
	Update(ctx context.Context, tenantID, id string, req *UpdateRuleRequest) (*AMLRule, error)
	ToggleStatus(ctx context.Context, tenantID, id string, enabled bool) (*AMLRule, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, tenantID string, req *CreateRuleRequest) (*AMLRule, error) {
	if req.RuleCode == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Field rule_code wajib diisi.")
	}
	if req.Name == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Field name wajib diisi.")
	}
	if req.RuleType == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Field rule_type wajib diisi.")
	}
	if req.Severity == "" {
		req.Severity = "MEDIUM"
	}
	ruleObj := &AMLRule{
		TenantID:      tenantID,
		RuleCode:      req.RuleCode,
		Name:          req.Name,
		Description:   req.Description,
		RuleType:      req.RuleType,
		Configuration: req.Configuration,
		Severity:      req.Severity,
		Enabled:       true,
		Version:       1,
	}
	if err := s.repo.Create(ctx, tenantID, ruleObj); err != nil {
		return nil, err
	}
	return ruleObj, nil
}

func (s *Service) GetByID(ctx context.Context, tenantID, id string) (*AMLRule, error) {
	if id == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Parameter ID rule wajib disertakan")
	}
	return s.repo.GetByID(ctx, tenantID, id)
}

func (s *Service) List(ctx context.Context, tenantID string, limit int, cursor string) (*RuleListResponse, error) {
	return s.repo.List(ctx, tenantID, limit, cursor)
}

func (s *Service) Update(ctx context.Context, tenantID, id string, req *UpdateRuleRequest) (*AMLRule, error) {
	if id == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Parameter ID rule wajib disertakan")
	}
	if req.Name == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Field name wajib diisi")
	}
	return s.repo.Update(ctx, tenantID, id, req)
}

func (s *Service) ToggleStatus(ctx context.Context, tenantID, id string, enabled bool) (*AMLRule, error) {
	if id == "" {
		return nil, errors.BadRequest(errors.CodeValidationError, "Parameter ID rule wajib disertakan")
	}
	return s.repo.ToggleStatus(ctx, tenantID, id, enabled)
}
