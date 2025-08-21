package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type CourseRepository interface {
	Create(ctx context.Context, ID string, c *requests.Course) error
	SetCreators(ctx context.Context, ID string, creators []string) error
	GetCreators(ctx context.Context, ID string) ([]models.CourseCreator, error)
	GetByID(ctx context.Context, ID string) (*Course, error)
	GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string, show string) ([]models.Course, error)
	Count(ctx context.Context, search string, show string) (int, error)
	UpdateByID(ctx context.Context, ID string, c *requests.Course) error
	DeleteByID(ctx context.Context, ID string) error
}

type Course struct {
	ID         string                 `db:"id"`
	Name       string                 `db:"name"`
	Creators   []models.CourseCreator `db:"creators"`
	IsArchived bool                   `db:"is_archived"`
	models.RecordStatus
}

func (c *Course) Model() *models.Course {
	return &models.Course{
		ID:   c.ID,
		Name: c.Name,
		RecordStatus: models.RecordStatus{
			IsDeleted: c.IsDeleted,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			DeletedAt: c.DeletedAt,
		},
	}
}
