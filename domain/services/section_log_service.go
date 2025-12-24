package services

import (
	"context"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/google/uuid"
)

type sectionLogService struct {
	sectionLogRepo      repositories.SectionLogRepository
	allowedFilterFields map[string]bool
}

type SectionLogService interface {
	Create(ctx context.Context, sectionID string, action string) error
	GetPaginationBySectionID(ctx context.Context, sectionID string, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.SectionLog, error)
	CountBySectionID(ctx context.Context, sectionID string, search string, filterParams map[string]string) (int, error)
}

func NewSectionLogService(sectionLogRepo repositories.SectionLogRepository) SectionLogService {
	return &sectionLogService{
		sectionLogRepo: sectionLogRepo,
		allowedFilterFields: map[string]bool{
			"action":     true,
			"ip_address": true,
			"created_at": true,
		},
	}
}

func (s *sectionLogService) Create(ctx context.Context, sectionID string, action string) error {
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	return s.sectionLogRepo.Create(ctx, id.String(), sectionID, action)
}

func (s *sectionLogService) GetPaginationBySectionID(ctx context.Context, sectionID string, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.SectionLog, error) {
	allowedSortFields := map[string]bool{
		"name":      true,
		"type":      true,
		"timestamp": true,
	}
	sanitizedSortBy, err := sanitize.SortBy(sortBy, allowedSortFields)
	if err != nil {
		return nil, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid sort by field",
			})
	}

	sanitizedSortOrder, err := sanitize.SortOrder(sortOrder)
	if err != nil {
		return nil, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid sort order",
			})
	}

	filters, err := sanitize.Filters(filterParams, s.allowedFilterFields)
	if err != nil {
		return nil, err
	}

	return s.sectionLogRepo.GetPaginationBySectionID(ctx, sectionID, page, limit, search, sanitizedSortBy, sanitizedSortOrder, filters)
}

func (s *sectionLogService) CountBySectionID(ctx context.Context, sectionID string, search string, filterParams map[string]string) (int, error) {
	filters, err := sanitize.Filters(filterParams, s.allowedFilterFields)
	if err != nil {
		return 0, err
	}
	return s.sectionLogRepo.CountBySectionID(ctx, sectionID, search, filters)
}
