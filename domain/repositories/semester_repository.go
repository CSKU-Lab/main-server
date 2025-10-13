package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

type SemesterRepository interface {
	Create(ctx context.Context, ID string, sem *requests.CreateSemester) error
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter) ([]models.Semester, error)
	Count(ctx context.Context, search string, filters []sanitize.Filter) (int, error)
	GetByID(ctx context.Context, ID string) (*models.Semester, error)
	UpdateByID(ctx context.Context, ID string, sem *requests.UpdateSemester) error
	DeleteByID(ctx context.Context, ID string) error
}
