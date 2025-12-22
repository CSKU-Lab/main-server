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

type LabService interface {
	GetByID(ctx context.Context, labID string) (*models.Lab, error)
	Create(ctx context.Context, req *requests.CreateLab, userID string) (string, error)
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.Lab, error)
	Count(ctx context.Context, search string, filterParams map[string]string) (int, error)
	UpdateByID(ctx context.Context, labID string, userID string, req *requests.BaseUpdateLab) error
	DeleteByID(ctx context.Context, labID string, userID string) error
}

type labService struct {
	labRepo             repositories.LabRepository
	courseRepo          repositories.CourseRepository
	uowRepo             repositories.UoWRepository
	allowedFilterFields map[string]bool
	allowedSortFields   map[string]bool
}

func NewLabService(
	labRepo repositories.LabRepository,
	courseRepo repositories.CourseRepository,
	uowRepo repositories.UoWRepository,
) LabService {
	return &labService{
		labRepo:    labRepo,
		courseRepo: courseRepo,
		uowRepo:    uowRepo,
		allowedFilterFields: map[string]bool{
			"display_name": true,
			"course_id":    true,
		},
		allowedSortFields: map[string]bool{
			"display_name": true,
			"created_at":   true,
			"updated_at":   true,
		},
	}
}

func (l *labService) GetByID(ctx context.Context, labID string) (*models.Lab, error) {
	lab, err := l.labRepo.GetByID(ctx, labID)
	if err != nil {
		return nil, err
	}
	return lab, nil
}

func (l *labService) Create(ctx context.Context, req *requests.CreateLab, userID string) (string, error) {
	_, err := l.courseRepo.GetByID(ctx, req.CourseID)
	if err != nil {
		return "", err
	}

	var labID string
	err = l.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		id, err := uuid.NewV7()
		if err != nil {
			return cserrors.New(&cserrors.Option{
				Message:    "Failed to generate UUID",
				HttpStatus: http.StatusInternalServerError,
			})
		}

		err = u.Lab().Create(ctx, id.String(), req, userID)
		if err != nil {
			return err
		}
		labID = id.String()

		return nil
	})
	if err != nil {
		return "", err
	}

	return labID, nil
}

func (l *labService) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.Lab, error) {
	sanitizedSortBy, err := sanitize.SortBy(sortBy, l.allowedSortFields)
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

	sanitizedFilters, err := sanitize.Filters(filterParams, l.allowedFilterFields)
	if err != nil {
		return nil, err
	}

	labs, err := l.labRepo.GetPagination(ctx, page, limit, search, sanitizedSortBy, sanitizedSortOrder, sanitizedFilters)
	if err != nil {
		return nil, err
	}

	return labs, nil
}

func (l *labService) Count(ctx context.Context, search string, filterParams map[string]string) (int, error) {
	filters, err := sanitize.Filters(filterParams, l.allowedFilterFields)
	if err != nil {
		return 0, err
	}

	return l.labRepo.Count(ctx, search, filters)
}

func (l *labService) UpdateByID(ctx context.Context, labID string, userID string, req *requests.BaseUpdateLab) error {
	lab, err := l.labRepo.GetByID(ctx, labID)
	if err != nil {
		return err
	}
	course, err := l.courseRepo.GetByID(ctx, lab.CourseID)
	if err != nil {
		return err
	}
	err = l.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		user, err := u.User().GetByID(ctx, userID)
		if err != nil {
			return err
		}

		for _, role := range user.Roles {
			if role == string(models.ADMIN) {
				return nil
			}
		}

		for _, creator := range course.Creators {
			if creator.ID == userID {
				return nil
			}
		}

		return cserrors.New(&cserrors.Option{
			Message:    "No Permission",
			HttpStatus: http.StatusForbidden,
		})
	})
	if err != nil {
		return err
	}

	err = l.labRepo.UpdateByID(ctx, labID, req)
	if err != nil {
		return err
	}

	return nil
}

func (l *labService) DeleteByID(ctx context.Context, labID string, userID string) error {
	lab, err := l.labRepo.GetByID(ctx, labID)
	if err != nil {
		return err
	}
	course, err := l.courseRepo.GetByID(ctx, lab.CourseID)
	if err != nil {
		return err
	}
	err = l.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		user, err := u.User().GetByID(ctx, userID)
		if err != nil {
			return err
		}

		for _, role := range user.Roles {
			if role == string(models.ADMIN) {
				return nil
			}
		}

		for _, creator := range course.Creators {
			if creator.ID == userID {
				return nil
			}
		}

		return cserrors.New(&cserrors.Option{
			Message:    "No Permission",
			HttpStatus: http.StatusForbidden,
		})
	})
	if err != nil {
		return err
	}

	err = l.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		labSections, err := u.LabSection().GetByLabID(ctx, labID)
		if err != nil {
			return err
		}

		for _, labSection := range labSections {
			data, err := u.LabSection().GetByID(ctx, labID, labSection.SectionID)
			if err != nil {
				return err
			}

			err = u.LabSection().ShiftUpPositions(ctx, labSection.SectionID, labID, labSection.Position)
			if err != nil {
				return err
			}

			err = u.LabSection().DeleteByID(ctx, data.ID)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	err = l.labRepo.DeleteByID(ctx, labID)
	if err != nil {
		return err
	}

	return nil
}
