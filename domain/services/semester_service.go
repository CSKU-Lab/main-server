package services

import (
	"context"
	"net/http"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/google/uuid"
)

type RawSection struct {
	ID         string
	Name       string
	Banner     *string
	CourseID   string
	SemesterID string
	CreatedAt  string
	UpdatedAt  string
}

type SemesterService interface {
	Create(ctx context.Context, sem *requests.CreateSemester) error
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.Semester, error)
	GetAffectedSections(ctx context.Context, semesterID string) ([]models.AffectedSection, error)
	GetByID(ctx context.Context, ID string) (*models.Semester, error)
	Count(ctx context.Context, search string, filterParams map[string]string) (int, error)
	UpdateByID(ctx context.Context, ID string, sem *requests.UpdateSemester) error
	DeleteByID(ctx context.Context, ID string) error
}

type semesterService struct {
	repo                repositories.SemesterRepository
	sectionRepo         repositories.SectionRepository
	courseRepo          repositories.CourseRepository
	allowedFilterFields map[string]bool
}

func NewSemesterService(repo repositories.SemesterRepository, sectionRepo repositories.SectionRepository, courseRepo repositories.CourseRepository) *semesterService {
	return &semesterService{
		repo:        repo,
		sectionRepo: sectionRepo,
		courseRepo:  courseRepo,
		allowedFilterFields: map[string]bool{
			"name":         true,
			"type":         true,
			"started_date": true,
		},
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

func (s *semesterService) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.Semester, error) {
	allowedSortFields := map[string]bool{
		"name":         true,
		"type":         true,
		"started_date": true,
	}
	sanitizedSortBy, err := sanitize.SortBy(sortBy, allowedSortFields)
	if err != nil {
		return nil, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid sort by field",
			})
	}

	sanitizedSortOrder, err := sanitize.SortOrder(sortOrder)
	if err != nil {
		return nil, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid sort order",
			})
	}

	if t, ok := filterParams["type__is"]; ok {
		t = strings.ToLower(t)
		if t != "first" && t != "second" && t != "summer" {
			return nil, cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid semester type filter",
			})
		}
		filterParams["type__is"] = strings.ToLower(t)
	}

	filters, err := sanitize.Filters(filterParams, s.allowedFilterFields)
	if err != nil {
		return nil, err
	}

	return s.repo.GetPagination(ctx, page, limit, search, sanitizedSortBy, sanitizedSortOrder, filters)

}

func (s *semesterService) Count(ctx context.Context, search string, filterParams map[string]string) (int, error) {
	filters, err := sanitize.Filters(filterParams, s.allowedFilterFields)
	if err != nil {
		return 0, err
	}

	return s.repo.Count(ctx, search, filters)
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

func (s *semesterService) GetAffectedSections(ctx context.Context, semesterID string) ([]models.AffectedSection, error) {
	sections, err := s.sectionRepo.GetRawBySemesterID(ctx, semesterID)
	if err != nil {
		return nil, err
	}

	courseWithSectionsMap := make(map[string][]string)
	for _, section := range sections {
		courseWithSectionsMap[section.CourseID] = append(courseWithSectionsMap[section.CourseID], section.Name)
	}

	courseWithSectionsSlice := make([]models.AffectedSection, 0, len(courseWithSectionsMap))
	for courseID, sectionNames := range courseWithSectionsMap {
		course, err := s.courseRepo.GetByID(ctx, courseID)
		if err != nil {
			return nil, err
		}
		courseWithSectionsSlice = append(courseWithSectionsSlice, models.AffectedSection{
			CourseName: course.Name,
			Sections:   sectionNames,
		})
	}

	return courseWithSectionsSlice, nil
}
