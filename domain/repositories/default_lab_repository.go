package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

type DefaultLabRepository interface {
	Create(ctx context.Context, req *requests.SetDefaultLab, id string, courseID string, labName string) error
	Update(ctx context.Context, req *requests.UpdateDefaultLab, id string) error
	GetByID(ctx context.Context, labID string, courseID string) (*models.DefaultLab, error)
	GetByLabID(ctx context.Context, labID string) ([]models.DefaultLab, error)
	DeleteByID(ctx context.Context, id string) error
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filterParams []sanitize.Filter) ([]models.DefaultLab, error)
	Count(ctx context.Context, filterParams []sanitize.Filter) (int, error)
	GetMaxPosition(ctx context.Context, courseID string, labID string) (int, error)
	ShiftDownPositions(ctx context.Context, courseID string, position int) error
	ShiftUpPositions(ctx context.Context, courseID string, labID string, position int) error
	ShiftInsertedPositions(ctx context.Context, courseID string, currPos int, reqPos int) error
	GetByCourseID(ctx context.Context, courseID string) ([]models.DefaultLab, error)
}
