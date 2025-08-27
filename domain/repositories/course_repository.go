package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type CourseRepository interface {
	Create(ctx context.Context, ID string, c *requests.Course) error
	GetByID(ctx context.Context, ID string) (*models.Course, error)
	GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string, show string) ([]models.Course, error)
	Count(ctx context.Context, search string, show string) (int, error)
	UpdateByID(ctx context.Context, ID string, c *requests.Course) error
	DeleteByID(ctx context.Context, ID string) error
}
