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

type LabMaterialService interface {
	Create(ctx context.Context, req *requests.SetLabMaterial, userID string) error
	DeleteByID(ctx context.Context, labID string, materialID string, userID string) error
	GetPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filterParams map[string]string) ([]models.LabMaterial, error)
	Count(ctx context.Context, filterParams map[string]string) (int, error)
}

type labMaterialService struct {
	labMaterialRepo     repositories.LabMaterialRepository
	uowRepo             repositories.UoWRepository
	allowedFilterFields map[string]bool
	allowedSortFields   map[string]bool
}

func NewLabMaterialService(labMaterialRepo repositories.LabMaterialRepository, uowRepo repositories.UoWRepository) LabMaterialService {
	return &labMaterialService{
		labMaterialRepo: labMaterialRepo,
		uowRepo:         uowRepo,
		allowedFilterFields: map[string]bool{
			"lab_id":      true,
			"material_id": true,
		},
		allowedSortFields: map[string]bool{
			"created_at": true,
		},
	}
}

func (lm *labMaterialService) mutationPermission(ctx context.Context, userID string, labID string) error {
	err := lm.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		user, err := u.User().GetByID(ctx, userID)
		if err != nil {
			return err
		}

		for _, role := range user.Roles {
			if role == string(models.ADMIN) {
				return nil
			}
		}

		lab, err := u.Lab().GetByID(ctx, labID)
		if err != nil {
			return err
		}

		courseCreator, err := u.CourseCreator().GetCreators(ctx, lab.CourseID)
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

func (lm *labMaterialService) rowExists(ctx context.Context, labID string, materialID string, labMaterialRepo repositories.LabMaterialRepository) (*models.LabMaterial, error) {
	err := lm.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		_, err := u.Lab().GetByID(ctx, labID)
		if err != nil {
			return err
		}

		_, err = u.Material().GetByID(ctx, materialID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	labMaterial, err := labMaterialRepo.GetByID(ctx, labID, materialID)
	if err != nil {
		return nil, err
	}
	return labMaterial, nil
}

func (lm *labMaterialService) Create(ctx context.Context, req *requests.SetLabMaterial, userID string) error {
	labMaterial, err := lm.rowExists(ctx, req.LabID, req.MaterialID, lm.labMaterialRepo)
	if err != nil && labMaterial != nil {
		return err
	}
	if labMaterial != nil {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusConflict,
			Message:    "Item already exists",
		})
	}

	err = lm.mutationPermission(ctx, userID, req.LabID)
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

	return lm.labMaterialRepo.Create(ctx, req, ID.String())
}

func (lm *labMaterialService) DeleteByID(ctx context.Context, labID string, materialID string, userID string) error {
	labMaterial, err := lm.rowExists(ctx, labID, materialID, lm.labMaterialRepo)
	if err != nil {
		return err
	}

	err = lm.mutationPermission(ctx, userID, labID)
	if err != nil {
		return err
	}

	err = lm.labMaterialRepo.DeleteByID(ctx, labMaterial.ID)
	if err != nil {
		return err
	}
	return nil
}

func (lm *labMaterialService) GetPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filterParams map[string]string) ([]models.LabMaterial, error) {
	sanitizedSortBy, err := sanitize.SortBy(sortBy, lm.allowedSortFields)
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

	sanitizedFilters, err := sanitize.Filters(filterParams, lm.allowedFilterFields)
	if err != nil {
		return nil, err
	}

	labMaterials, err := lm.labMaterialRepo.GetPagination(ctx, page, limit, sanitizedSortBy, sanitizedSortOrder, sanitizedFilters)
	if err != nil {
		return nil, err
	}

	return labMaterials, nil
}

func (lm *labMaterialService) Count(ctx context.Context, filterParams map[string]string) (int, error) {
	filters, err := sanitize.Filters(filterParams, lm.allowedFilterFields)
	if err != nil {
		return 0, err
	}

	return lm.labMaterialRepo.Count(ctx, filters)
}
