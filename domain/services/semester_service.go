package services

import (
	"context"

	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/repositories"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
	"github.com/google/uuid"
)

type SemesterService interface {
	Create(ctx context.Context, sem *requests.Semester) (*models.Semester, error)
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.Semester, error)
	Count(ctx context.Context, search string) (int, error)
	GetByID(ctx context.Context, ID string) (*models.Semester, error)
	UpdateByID(ctx context.Context, ID string, sem *requests.Semester) (*models.Semester, error)
	DeleteByID(ctx context.Context, ID string) error
}

type semesterService struct {
	repo repositories.SemesterRepository
}

func NewSemesterService(repo repositories.SemesterRepository) *semesterService {
	return &semesterService{
		repo: repo,
	}
}

func (s *semesterService) Create(ctx context.Context, sem *requests.Semester) (*models.Semester, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, cserrors.New(cserrors.INTERNAL_SERVER_ERROR, "Cannot generate user ID")
	}

	return s.repo.Create(ctx, id.String(), sem)
}

func (s *semesterService) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.Semester, error) {
	sanitizedSortBy, err := sanitizeSortBy(sortBy, &models.User{})
	if err != nil {
		return nil, err
	}

	sanitizedSortOrder := sanitizeSortOrder(sortOrder)

	return s.repo.GetPagination(ctx, page, limit, search, sanitizedSortBy, sanitizedSortOrder)

}

func (s *semesterService) Count(ctx context.Context, search string) (int, error) {
	return s.repo.Count(ctx, search)
}

func (s *semesterService) GetByID(ctx context.Context, ID string) (*models.Semester, error) {
	return s.repo.GetByID(ctx, ID)
}

func (s *semesterService) UpdateByID(ctx context.Context, ID string, sem *requests.Semester) (*models.Semester, error) {
	return s.repo.UpdateByID(ctx, ID, sem)
}

func (s *semesterService) DeleteByID(ctx context.Context, ID string) error {
	return s.repo.DeleteByID(ctx, ID)
}
