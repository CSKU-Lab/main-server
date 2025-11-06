package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

type MaterialRepository interface {
	Create(ctx context.Context, ID string, createdByUserID string, req *requests.CreateMaterial) error
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter) ([]models.Material, error)
	Count(ctx context.Context, search string, filters []sanitize.Filter) (int, error)
	GetByID(ctx context.Context, ID string) (*models.Material, error)
	UpdateByID(ctx context.Context, ID string, req *requests.UpdateMaterial) error
	DeleteByID(ctx context.Context, ID string) error
}
