package repositories

import (
	"context"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
)

type SemesterRepository interface {
	Create(ctx context.Context, ID string, sem *requests.Semester) (*models.Semester, error)
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.Semester, error)
	Count(ctx context.Context, search string) (int, error)
	GetByID(ctx context.Context, ID string) (*models.Semester, error)
	UpdateByID(ctx context.Context, ID string, sem *requests.Semester) (*models.Semester, error)
	DeleteByID(ctx context.Context, ID string) error
}
