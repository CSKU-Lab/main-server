package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

type LabMaterialRepository interface {
	Create(ctx context.Context, req *requests.SetLabMaterial, id string) error
	GetByID(ctx context.Context, labID string, materilaID string) (*models.LabMaterial, error)
	DeleteByID(ctx context.Context, id string) error
	GetPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filterParams []sanitize.Filter) ([]models.LabMaterial, error)
	Count(ctx context.Context, filterParams []sanitize.Filter) (int, error)
}
