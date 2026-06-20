package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type CourseRepository interface {
	Create(ctx context.Context, ID string, c *requests.CreateCourse) error
	GetByID(ctx context.Context, ID string) (*models.Course, error)
	GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string, show string, visibility string) ([]models.Course, error)
	Count(ctx context.Context, search string, show string, visibility string) (int, error)
	GetFeatured(ctx context.Context, limit int) ([]models.Course, error)
	GetPaginationForStudent(ctx context.Context, studentID string, page int, pageSize int, search string, sortBy string, sortOrder string, show string) ([]models.Course, error)
	CountForStudent(ctx context.Context, studentID string, search string, show string) (int, error)
	GetPaginationForInstructor(ctx context.Context, instructorID string, page int, pageSize int, search string, sortBy string, sortOrder string, show string) ([]models.Course, error)
	CountForInstructor(ctx context.Context, instructorID string, search string, show string) (int, error)
	UpdateByID(ctx context.Context, ID string, c *requests.UpdateCourse) error
	DeleteByID(ctx context.Context, ID string) error
}
