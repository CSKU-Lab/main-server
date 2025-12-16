package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

type LabSectionRepository interface {
	ShiftUpPositions(ctx context.Context, sectionID string, position int) error
	ShiftDownPositions(ctx context.Context, sectionID string, position int) error
	Create(ctx context.Context, req *requests.SetLabSection, id string) error
	GetMaxPosition(ctx context.Context, sectionID string, labID string) (int, error)
	GetPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filters []sanitize.Filter) ([]models.LabSection, error)
	GetByID(ctx context.Context, labID string, sectionID string) (*models.LabSection, error)
	UpdateByID(ctx context.Context, labID string, sectionID string, id string, req *requests.UpdateLabSection) error
	DeleteByID(ctx context.Context, id string) error
	Count(ctx context.Context, filters []sanitize.Filter) (int, error)
}
