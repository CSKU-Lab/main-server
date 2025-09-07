package services

import (
	"context"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/google/uuid"
)

type SemesterService interface {
	Create(ctx context.Context, sem *requests.CreateSemester) error
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.Semester, error)
	GetByID(ctx context.Context, ID string) (*models.Semester, error)
	Count(ctx context.Context, search string) (int, error)
	UpdateByID(ctx context.Context, ID string, sem *requests.UpdateSemester) error
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

func (s *semesterService) Create(ctx context.Context, sem *requests.CreateSemester) error {
	id, err := uuid.NewV7()
	if err != nil {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Cannot generate user ID",
		})
	}

	return s.repo.Create(ctx, id.String(), sem)
}

func (s *semesterService) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string) ([]models.Semester, error) {
	sanitizedSortBy, err := sanitizeSortBy(sortBy, []string{"name", "type", "started_date"})
	if err != nil {
		return nil, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid sort by field",
			})
	}

	sanitizedSortOrder, err := sanitizeSortOrder(sortOrder)
	if err != nil {
		return nil, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid sort order",
			})
	}

	return s.repo.GetPagination(ctx, page, limit, search, sanitizedSortBy, sanitizedSortOrder)

}

func (s *semesterService) Count(ctx context.Context, search string) (int, error) {
	return s.repo.Count(ctx, search)
}

func (s *semesterService) GetByID(ctx context.Context, ID string) (*models.Semester, error) {
	return s.repo.GetByID(ctx, ID)
}

func (s *semesterService) UpdateByID(ctx context.Context, ID string, sem *requests.UpdateSemester) error {
	return s.repo.UpdateByID(ctx, ID, sem)
}

func (s *semesterService) DeleteByID(ctx context.Context, ID string) error {
	return s.repo.DeleteByID(ctx, ID)
}
