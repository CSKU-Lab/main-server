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
	GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string, show string) ([]models.Course, error)
	Count(ctx context.Context, search string, show string) (int, error)
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

	err = s.repo.Create(ctx, id.String(), c)
	if err != nil {
		return nil, err
	}

	course, err := s.GetByID(ctx, id.String())
	if err != nil {
		return nil, err
	}

	return course, nil
}

func (s *courseService) GetByID(ctx context.Context, ID string) (*models.Course, error) {
	course, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	creators, err := s.repo.GetCreators(ctx, ID)
	if err != nil {
		return nil, err
	}

	courseModel := course.Model()
	courseModel.Creators = creators

	return courseModel, nil
}

func (s *courseService) GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string, show string) ([]models.Course, error) {
	sanitizedSortBy, err := sanitizeSortBy(sortBy, &models.Course{})
	if err != nil {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "Invalid sort by field")
	}

	sanitizedSortOrder, err := sanitizeSortOrder(sortOrder)
	if err != nil {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "Invalid sort order")
	}

	if show != "all" && show != "active" && show != "archived" {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "Invalid show value. Must be 'all', 'active', or 'archived'")
	}

	return s.repo.GetPagination(ctx, page, pageSize, search, sanitizedSortBy, sanitizedSortOrder, show)

}

func (s *courseService) Count(ctx context.Context, search string, show string) (int, error) {
	if show != "all" && show != "active" && show != "archived" {
		return 0, cserrors.New(cserrors.BAD_REQUEST, "Invalid show value. Must be 'all', 'active', or 'archived'")
	}

	return s.repo.Count(ctx, search, show)
}

func (s *courseService) UpdateByID(ctx context.Context, ID string, c *requests.Course) (*models.Course, error) {
	if c.Creators != nil && len(c.Creators) == 0 {
		return nil, cserrors.New(cserrors.BAD_REQUEST, "At least one creator is required")
	}

	if c.Creators != nil {
		err := s.repo.SetCreators(ctx, ID, c.Creators)
		if err != nil {
			return nil, err
		}
	}

	if c.Name != "" {
		err := s.repo.UpdateByID(ctx, ID, c)
		if err != nil {
			return nil, err
		}

	}

	return s.GetByID(ctx, ID)
}

func (s *courseService) DeleteByID(ctx context.Context, ID string) error {
	return s.repo.DeleteByID(ctx, ID)
}
