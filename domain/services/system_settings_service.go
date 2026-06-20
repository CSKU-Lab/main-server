package services

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/repositories"
)

const defaultCompareScriptIDKey = "default_compare_script_id"

type SystemSettingsService interface {
	GetDefaultCompareScriptID(ctx context.Context) string
	SetDefaultCompareScriptID(ctx context.Context, id string) error
}

type systemSettingsService struct {
	repo repositories.SystemSettingsRepository
}

func NewSystemSettingsService(repo repositories.SystemSettingsRepository) SystemSettingsService {
	return &systemSettingsService{repo: repo}
}

func (s *systemSettingsService) GetDefaultCompareScriptID(ctx context.Context) string {
	val, err := s.repo.Get(ctx, defaultCompareScriptIDKey)
	if err != nil || val == nil {
		return ""
	}
	return *val
}

func (s *systemSettingsService) SetDefaultCompareScriptID(ctx context.Context, id string) error {
	return s.repo.Set(ctx, defaultCompareScriptIDKey, id)
}
