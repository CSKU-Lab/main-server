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

type LabSectionService interface {
	Create(ctx context.Context, req *requests.SetLabSection, userID string, sectionID string) error
	GetPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filterParams map[string]string) ([]models.LabSection, error)
	Update(ctx context.Context, userID string, sectionID string, req *requests.UpdateLabSection) error
	Delete(ctx context.Context, sectionID string, userID string, req *requests.DeleteLabSection) error
	Count(ctx context.Context, filterParams map[string]string) (int, error)
}

type labSectionService struct {
	labSectionRepo      repositories.LabSectionRepository
	uowRepo             repositories.UoWRepository
	labRepo             repositories.LabRepository
	sectionRepo         repositories.SectionRepository
	allowedFilterFields map[string]bool
	allowedSortFields   map[string]bool
}

func NewLabSectionService(labSectionRepo repositories.LabSectionRepository, uowRepo repositories.UoWRepository, labRepo repositories.LabRepository, sectionRepo repositories.SectionRepository) LabSectionService {
	return &labSectionService{
		labSectionRepo: labSectionRepo,
		uowRepo:        uowRepo,
		labRepo:        labRepo,
		sectionRepo:    sectionRepo,
		allowedFilterFields: map[string]bool{
			"lab_id":     true,
			"section_id": true,
		},
		allowedSortFields: map[string]bool{
			"position": true,
		},
	}
}

func (ls *labSectionService) rowExists(ctx context.Context, labID string, sectionID string) error {
	_, err := ls.labRepo.GetByID(ctx, labID)
	if err != nil {
		return err
	}
	_, err = ls.sectionRepo.GetByID(ctx, sectionID)
	if err != nil {
		return err
	}
	return nil
}

func (ls *labSectionService) mutationPermission(ctx context.Context, userID string, sectionID string) error {
	err := ls.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		user, err := u.User().GetByID(ctx, userID)
		if err != nil {
			return err
		}

		for _, role := range user.Roles {
			if role == string(models.ADMIN) {
				return nil
			}
		}

		sectionInstructors, err := u.SectionInstructor().Get(ctx, sectionID)
		if err != nil {
			return err
		}

		section, err := u.Section().GetRawByID(ctx, sectionID)
		if err != nil {
			return err
		}

		courseCreator, err := u.CourseCreator().GetCreators(ctx, section.CourseID)
		if err != nil {
			return err
		}

		for _, creator := range courseCreator {
			if creator.ID == userID {
				return nil
			}
		}

		for _, instructor := range sectionInstructors {
			if instructor.ID == userID {
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

func (ls *labSectionService) rearrangeUpdatedIndex(ctx context.Context, sectionID string, position *int, labID string) error {
	err := ls.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		maxPos, err := ls.labSectionRepo.GetMaxPosition(ctx, sectionID, labID)
		if err != nil {
			return err
		}

		if maxPos < *position {
			*position = maxPos
		}

		err = ls.labSectionRepo.ShiftDownPositions(ctx, sectionID, *position)
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

func (ls *labSectionService) rearrangeDeletedIndex(ctx context.Context, sectionID string, labID string, position int) error {
	err := ls.labSectionRepo.ShiftUpPositions(ctx, sectionID, labID, position)
	if err != nil {
		return err
	}
	return nil
}

func (ls *labSectionService) Create(ctx context.Context, req *requests.SetLabSection, userID string, sectionID string) error {
	err := ls.rowExists(ctx, req.LabID, sectionID)
	if err != nil {
		return err
	}
	labSection, err := ls.labSectionRepo.GetByID(ctx, req.LabID, sectionID)
	if err != nil && labSection != nil {
		return err
	}
	if labSection != nil {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusConflict,
			Message:    "Item already exists",
		})
	}

	err = ls.mutationPermission(ctx, userID, sectionID)
	if err != nil {
		return err
	}

	err = ls.rearrangeUpdatedIndex(ctx, sectionID, &req.Position, req.LabID)
	if err != nil {
		return err
	}

	ID, err := uuid.NewV7()
	if err != nil {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Cannot generate uuid",
		})
	}

	err = ls.labSectionRepo.Create(ctx, req, ID.String(), sectionID)
	if err != nil {
		return err
	}

	return nil
}

func (ls *labSectionService) GetPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filterParams map[string]string) ([]models.LabSection, error) {
	sanitizedSortBy, err := sanitize.SortBy(sortBy, ls.allowedSortFields)
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

	sanitizedFilters, err := sanitize.Filters(filterParams, ls.allowedFilterFields)
	if err != nil {
		return nil, err
	}

	labSections, err := ls.labSectionRepo.GetPagination(ctx, page, limit, sanitizedSortBy, sanitizedSortOrder, sanitizedFilters)
	if err != nil {
		return nil, err
	}

	return labSections, nil
}

func (ls *labSectionService) Count(ctx context.Context, filterParams map[string]string) (int, error) {
	filters, err := sanitize.Filters(filterParams, ls.allowedFilterFields)
	if err != nil {
		return 0, err
	}

	return ls.labSectionRepo.Count(ctx, filters)
}

func (ls *labSectionService) Update(ctx context.Context, userID string, sectionID string, req *requests.UpdateLabSection) error {
	err := ls.rowExists(ctx, req.LabID, sectionID)
	if err != nil {
		return err
	}
	labSection, err := ls.labSectionRepo.GetByID(ctx, req.LabID, sectionID)
	if err != nil {
		return err
	}

	err = ls.mutationPermission(ctx, userID, sectionID)
	if err != nil {
		return err
	}

	err = ls.rearrangeUpdatedIndex(ctx, sectionID, &req.Position, req.LabID)
	if err != nil {
		return err
	}

	err = ls.labSectionRepo.UpdateByID(ctx, req.LabID, sectionID, labSection.ID, req)
	if err != nil {
		return err
	}
	return nil
}

func (ls *labSectionService) Delete(ctx context.Context, sectionID string, userID string, req *requests.DeleteLabSection) error {
	err := ls.rowExists(ctx, req.LabID, sectionID)
	if err != nil {
		return err
	}
	labSection, err := ls.labSectionRepo.GetByID(ctx, req.LabID, sectionID)
	if err != nil {
		return err
	}

	err = ls.mutationPermission(ctx, userID, sectionID)
	if err != nil {
		return err
	}

	err = ls.rearrangeDeletedIndex(ctx, sectionID, req.LabID, labSection.Position)
	if err != nil {
		return err
	}

	err = ls.labSectionRepo.DeleteByID(ctx, labSection.ID)
	if err != nil {
		return err
	}
	return nil
}
