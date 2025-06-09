package repositories

import (
	"context"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
)

type CourseRepository interface {
	Create(ctx context.Context, ID string, c *requests.Course) (*models.Course, error)
	GetByID(ctx context.Context, ID string) (*models.Course, error)
	GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string) ([]models.Course, error)
	Count(ctx context.Context, search string) (int, error)
	UpdateByID(ctx context.Context, ID string, c *requests.Course) (*models.Course, error)
	DeleteByID(ctx context.Context, ID string) error
}
