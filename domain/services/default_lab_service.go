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

type DefaultLabService interface {
	Create(ctx context.Context, req *requests.SetDefaultLab, userID string, courseID string) error
	Update(ctx context.Context, req *requests.UpdateDefaultLab, userID string, courseID string) error
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.DefaultLab, error)
	Delete(ctx context.Context, req *requests.DeleteDefaultLab, userID string, courseID string) error
	Count(ctx context.Context, filterParams map[string]string) (int, error)
}

type defaultLabService struct {
	defaultLabRepo      repositories.DefaultLabRepository
	uowRepo             repositories.UoWRepository
	courseRepo          repositories.CourseRepository
	labRepo             repositories.LabRepository
	allowedFilterFields map[string]bool
	allowedSortFields   map[string]bool
}

func NewDefaultLabService(defaultLabRepo repositories.DefaultLabRepository, uowRepo repositories.UoWRepository, courseRepo repositories.CourseRepository, labRepo repositories.LabRepository) DefaultLabService {
	return &defaultLabService{
		defaultLabRepo: defaultLabRepo,
		uowRepo:        uowRepo,
		courseRepo:     courseRepo,
		labRepo:        labRepo,
		allowedFilterFields: map[string]bool{
			"lab_id":    true,
			"course_id": true,
		},
		allowedSortFields: map[string]bool{
			"position":   true,
			"created_at": true,
		},
	}
}

func (dl *defaultLabService) mutationPermission(ctx context.Context, userID string, courseID string) error {
	err := dl.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		user, err := u.User().GetByID(ctx, userID)
		if err != nil {
			return err
		}

		for _, role := range user.Roles {
			if role == string(models.ADMIN) {
				return nil
			}
		}

		courseCreator, err := u.CourseCreator().GetCreators(ctx, courseID)
		if err != nil {
			return err
		}

		for _, creator := range courseCreator {
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
	return nil
}

func (dl *defaultLabService) rowExists(ctx context.Context, labID string, courseID string) error {
	_, err := dl.labRepo.GetByID(ctx, labID)
	if err != nil {
		return err
	}

	_, err = dl.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return err
	}
	return nil
}

func (dl *defaultLabService) rearrangeCreatedIndex(ctx context.Context, courseID string, labID string, position int) (int, error) {
	err := dl.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		maxPos, err := dl.defaultLabRepo.GetMaxPosition(ctx, courseID, labID)
		if err != nil {
			return err
		}

		if maxPos < position || position <= 0 {
			position = maxPos
		}

		err = dl.defaultLabRepo.ShiftDownPositions(ctx, courseID, position)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return position, nil
}

func (dl *defaultLabService) rearrangeUpdatedIndex(ctx context.Context, courseID string, labID string, currPos int, reqPos int) error {
	err := dl.defaultLabRepo.ShiftInsertedPositions(ctx, courseID, currPos, reqPos)
	if err != nil {
		return err
	}
	return nil
}

func (dl *defaultLabService) Create(ctx context.Context, req *requests.SetDefaultLab, userID string, courseID string) error {
	err := dl.rowExists(ctx, req.LabID, courseID)
	if err != nil {
		return err
	}

	err = dl.mutationPermission(ctx, userID, courseID)
	if err != nil {
		return err
	}

	lab, err := dl.labRepo.GetByID(ctx, req.LabID)
	if err != nil {
		return err
	}

	defaultLab, err := dl.defaultLabRepo.GetByID(ctx, req.LabID, courseID)
	if err != nil && defaultLab != nil {
		return err
	}
	if defaultLab != nil {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusConflict,
			Message:    "Item already exists",
		})
	}

	position := 0
	if req.Position != nil {
		position = *req.Position
	}
	index, err := dl.rearrangeCreatedIndex(ctx, courseID, req.LabID, position)
	if err != nil {
		return err
	}
	req.Position = &index

	ID, err := uuid.NewV7()
	if err != nil {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Cannot generate uuid",
		})
	}
	err = dl.defaultLabRepo.Create(ctx, req, ID.String(), courseID, lab.DisplayName)
	if err != nil {
		return err
	}

	err = dl.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		defaultValue := true
		err = u.Lab().UpdateByID(ctx, req.LabID, &requests.BaseUpdateLab{
			lab.DisplayName,
			lab.CourseID,
			&defaultValue,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (dl *defaultLabService) Update(ctx context.Context, req *requests.UpdateDefaultLab, userID string, courseID string) error {
	err := dl.rowExists(ctx, req.LabID, courseID)
	if err != nil {
		return err
	}

	err = dl.mutationPermission(ctx, userID, courseID)
	if err != nil {
		return err
	}

	_, err = dl.labRepo.GetByID(ctx, req.LabID)
	if err != nil {
		return err
	}

	defaultLab, err := dl.defaultLabRepo.GetByID(ctx, req.LabID, courseID)
	if err != nil {
		return err
	}

	reqPos := req.Position
	currPos := defaultLab.Position
	err = dl.rearrangeUpdatedIndex(ctx, courseID, req.LabID, currPos, reqPos)
	if err != nil {
		return err
	}

	err = dl.defaultLabRepo.Update(ctx, req, defaultLab.ID)
	if err != nil {
		return err
	}
	return nil
}

func (dl *defaultLabService) rearrangeDeletedIndex(ctx context.Context, courseID string, labID string, position int) error {
	err := dl.defaultLabRepo.ShiftUpPositions(ctx, courseID, labID, position)
	if err != nil {
		return err
	}
	return nil
}

func (dl *defaultLabService) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.DefaultLab, error) {
	sanitizedSortBy, err := sanitize.SortBy(sortBy, dl.allowedSortFields)
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

	sanitizedFilters, err := sanitize.Filters(filterParams, dl.allowedFilterFields)
	if err != nil {
		return nil, err
	}

	defaultLabs, err := dl.defaultLabRepo.GetPagination(ctx, page, limit, search, sanitizedSortBy, sanitizedSortOrder, sanitizedFilters)
	if err != nil {
		return nil, err
	}

	return defaultLabs, nil
}

func (dl *defaultLabService) Delete(ctx context.Context, req *requests.DeleteDefaultLab, userID string, courseID string) error {
	err := dl.rowExists(ctx, req.LabID, courseID)
	if err != nil {
		return err
	}

	err = dl.mutationPermission(ctx, userID, courseID)
	if err != nil {
		return err
	}

	defaultLab, err := dl.defaultLabRepo.GetByID(ctx, req.LabID, courseID)
	if err != nil {
		return err
	}

	err = dl.rearrangeDeletedIndex(ctx, courseID, req.LabID, defaultLab.Position)
	if err != nil {
		return err
	}

	err = dl.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		defaultValue := false
		err := u.Lab().UpdateByID(ctx, req.LabID, &requests.BaseUpdateLab{
			CourseID:  courseID,
			IsDefault: &defaultValue,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = dl.defaultLabRepo.DeleteByID(ctx, defaultLab.ID)
	if err != nil {
		return err
	}
	return nil
}

func (dl *defaultLabService) Count(ctx context.Context, filterParams map[string]string) (int, error) {
	filters, err := sanitize.Filters(filterParams, dl.allowedFilterFields)
	if err != nil {
		return 0, err
	}

	return dl.defaultLabRepo.Count(ctx, filters)
}
