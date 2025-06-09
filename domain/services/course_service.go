package services

import (
	"context"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/repositories"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
	"github.com/google/uuid"
)

type CourseService interface {
	Create(ctx context.Context, c *requests.Course, userID string) (*models.Course, error)
	GetByID(ctx context.Context, ID string) (*models.Course, error)
	GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string) ([]models.Course, error)
	Count(ctx context.Context, search string) (int, error)
	UpdateByID(ctx context.Context, ID string, c *requests.Course) (*models.Course, error)
	DeleteByID(ctx context.Context, ID string) error
}

type courseService struct {
	repo repositories.CourseRepository
}

func NewCourseService(repo repositories.CourseRepository) CourseService {
	return &courseService{
		repo: repo,
	}
}

func (s *courseService) Create(ctx context.Context, c *requests.Course, userID string) (*models.Course, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Cannot generate user ID")
	}

	return s.repo.Create(ctx, id.String(), c)
}

func (s *courseService) GetByID(ctx context.Context, ID string) (*models.Course, error) {
	return s.repo.GetByID(ctx, ID)
}

func (s *courseService) GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string) ([]models.Course, error) {
	sanitizedSortBy, err := sanitizeSortBy(sortBy, &models.Course{})
	if err != nil {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "Invalid sort by field")
	}

	sanitizedSortOrder, err := sanitizeSortOrder(sortOrder)
	if err != nil {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "Invalid sort order")
	}

	return s.repo.GetPagination(ctx, page, pageSize, search, sanitizedSortBy, sanitizedSortOrder)

}

func (s *courseService) Count(ctx context.Context, search string) (int, error) {
	return s.repo.Count(ctx, search)
}

func (s *courseService) UpdateByID(ctx context.Context, ID string, c *requests.Course) (*models.Course, error) {
	return s.repo.UpdateByID(ctx, ID, c)
}

func (s *courseService) DeleteByID(ctx context.Context, ID string) error {
	return s.repo.DeleteByID(ctx, ID)
}
