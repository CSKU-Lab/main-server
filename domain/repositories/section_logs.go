package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

type SectionLogRepository interface {
	Create(ctx context.Context, id string, sectionID string, action string) error
	GetPaginationBySectionID(ctx context.Context, sectionID string, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter) ([]models.SectionLog, error)
	CountBySectionID(ctx context.Context, sectionID string, search string, filters []sanitize.Filter) (int, error)
}
