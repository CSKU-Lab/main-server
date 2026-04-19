package services

import (
	"context"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/google/uuid"
)

type CourseService interface {
	Create(ctx context.Context, c *requests.CreateCourse) (*models.Course, error)
	GetByID(ctx context.Context, ID string) (*models.Course, error)
	GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string, show string, visibility string) ([]models.Course, error)
	Count(ctx context.Context, search string, show string, visibility string) (int, error)
	GetFeatured(ctx context.Context, limit int) ([]models.Course, error)
	GetPaginationForStudent(ctx context.Context, studentID string, page int, pageSize int, search string, sortBy string, sortOrder string, show string) ([]models.Course, error)
	CountForStudent(ctx context.Context, studentID string, search string, show string) (int, error)
	UpdateByID(ctx context.Context, ID string, c *requests.UpdateCourse) error
	DeleteByID(ctx context.Context, ID string) error
}

type courseService struct {
	courseRepo        repositories.CourseRepository
	courseCreatorRepo repositories.CourseCreatorRepository
	uowRepo           repositories.UoWRepository
}

func NewCourseService(courseRepo repositories.CourseRepository, courseCreatorRepo repositories.CourseCreatorRepository, uowRepo repositories.UoWRepository) CourseService {
	return &courseService{
		courseRepo:        courseRepo,
		courseCreatorRepo: courseCreatorRepo,
		uowRepo:           uowRepo,
	}
}

func (s *courseService) Create(ctx context.Context, c *requests.CreateCourse) (*models.Course, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Cannot generate course ID",
			})
	}

	err = s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		err := u.Course().Create(ctx, id.String(), c)
		if err != nil {
			return err
		}

		err = u.CourseCreator().SetCreators(ctx, id.String(), c.Creators)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id.String())
}

func (s *courseService) GetByID(ctx context.Context, ID string) (*models.Course, error) {
	course, err := s.courseRepo.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	creators, err := s.courseCreatorRepo.GetCreators(ctx, ID)
	if err != nil {
		return nil, err
	}

	course.Creators = creators

	return course, nil
}

func (s *courseService) GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string, show string, visibility string) ([]models.Course, error) {
	allowedSortFields := map[string]bool{
		"name":       true,
		"type":       true,
		"created_at": true,
	}
	sanitizedSortBy, err := sanitize.SortBy(sortBy, allowedSortFields)
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Invalid sort by field",
		})
	}

	sanitizedSortOrder, err := sanitize.SortOrder(sortOrder)
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Invalid sort order",
		})
	}

	if show != "all" && show != "active" && show != "archived" {
		return nil, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid show value. Must be 'all', 'active', or 'archived'",
			})
	}

	courses, err := s.courseRepo.GetPagination(ctx, page, pageSize, search, sanitizedSortBy, sanitizedSortOrder, show, visibility)
	if err != nil {
		return nil, err
	}

	for i, course := range courses {
		creators, err := s.courseCreatorRepo.GetCreators(ctx, course.ID)
		if err != nil {
			return nil, err
		}
		courses[i].Creators = creators
	}

	return courses, nil
}

func (s *courseService) Count(ctx context.Context, search string, show string, visibility string) (int, error) {
	if show != "all" && show != "active" && show != "archived" {
		return 0, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid show value. Must be 'all', 'active', or 'archived'",
			})
	}

	return s.courseRepo.Count(ctx, search, show, visibility)
}

func (s *courseService) GetFeatured(ctx context.Context, limit int) ([]models.Course, error) {
	courses, err := s.courseRepo.GetFeatured(ctx, limit)
	if err != nil {
		return nil, err
	}

	for i, course := range courses {
		creators, err := s.courseCreatorRepo.GetCreators(ctx, course.ID)
		if err != nil {
			return nil, err
		}
		courses[i].Creators = creators
	}

	return courses, nil
}

func (s *courseService) GetPaginationForStudent(ctx context.Context, studentID string, page int, pageSize int, search string, sortBy string, sortOrder string, show string) ([]models.Course, error) {
	allowedSortFields := map[string]bool{
		"name":       true,
		"type":       true,
		"created_at": true,
	}
	sanitizedSortBy, err := sanitize.SortBy(sortBy, allowedSortFields)
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid sort by field"})
	}

	sanitizedSortOrder, err := sanitize.SortOrder(sortOrder)
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid sort order"})
	}

	if show != "all" && show != "active" && show != "archived" {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid show value. Must be 'all', 'active', or 'archived'"})
	}

	courses, err := s.courseRepo.GetPaginationForStudent(ctx, studentID, page, pageSize, search, sanitizedSortBy, sanitizedSortOrder, show)
	if err != nil {
		return nil, err
	}

	for i, course := range courses {
		creators, err := s.courseCreatorRepo.GetCreators(ctx, course.ID)
		if err != nil {
			return nil, err
		}
		courses[i].Creators = creators
	}

	return courses, nil
}

func (s *courseService) CountForStudent(ctx context.Context, studentID string, search string, show string) (int, error) {
	if show != "all" && show != "active" && show != "archived" {
		return 0, cserrors.New(&cserrors.Option{HttpStatus: http.StatusBadRequest, Message: "Invalid show value. Must be 'all', 'active', or 'archived'"})
	}

	return s.courseRepo.CountForStudent(ctx, studentID, search, show)
}

func (s *courseService) UpdateByID(ctx context.Context, ID string, c *requests.UpdateCourse) error {
	if c.Name == nil && c.Creators == nil && c.Visibility == nil {
		return nil
	}

	return s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		if c.Creators != nil {
			err := u.CourseCreator().SetCreators(ctx, ID, *c.Creators)
			if err != nil {
				return err
			}
		}

		return u.Course().UpdateByID(ctx, ID, c)
	})
}

func (s *courseService) DeleteByID(ctx context.Context, ID string) error {
	return s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		err := u.Course().DeleteByID(ctx, ID)
		if err != nil {
			return err
		}

		err = u.Section().DeleteByCourseID(ctx, ID)
		if err != nil {
			return err
		}

		return nil
	})
}
