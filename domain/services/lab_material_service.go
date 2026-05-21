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
	Create(ctx context.Context, req *requests.SetLabMaterial, userID string, labID string) error
	Delete(ctx context.Context, labID string, userID string, req *requests.DeleteLabMaterial) error
	GetPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filterParams map[string]string) ([]models.LabMaterial, error)
	GetByLabID(ctx context.Context, labID string) ([]models.LabMaterial, error)
	GetByLabAndMaterialID(ctx context.Context, labID string, materialID string) (*models.LabMaterial, error)
	Count(ctx context.Context, filterParams map[string]string) (int, error)
	UpdatePosition(ctx context.Context, labID string, materialID string, userID string, position int) error
}

type labMaterialService struct {
	labMaterialRepo     repositories.LabMaterialRepository
	uowRepo             repositories.UoWRepository
	labRepo             repositories.LabRepository
	materialRepo        repositories.MaterialRepository
	readMaterialTagRepo repositories.ReadMaterialTagRepository
	userRepo            repositories.User
	allowedFilterFields map[string]bool
	allowedSortFields   map[string]bool
}

func NewLabMaterialService(labMaterialRepo repositories.LabMaterialRepository, uowRepo repositories.UoWRepository, labRepo repositories.LabRepository, materialRepo repositories.MaterialRepository, readMaterialTagRepo repositories.ReadMaterialTagRepository, userRepo repositories.User) LabMaterialService {
	return &labMaterialService{
		labMaterialRepo:     labMaterialRepo,
		uowRepo:             uowRepo,
		labRepo:             labRepo,
		materialRepo:        materialRepo,
		readMaterialTagRepo: readMaterialTagRepo,
		userRepo:            userRepo,
		allowedFilterFields: map[string]bool{
			"lab_id":      true,
			"material_id": true,
		},
		allowedSortFields: map[string]bool{
			"created_at": true,
			"position":   true,
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

func (lm *labMaterialService) rowExists(ctx context.Context, labID string, materialID string) error {
	lab, err := lm.labRepo.GetByID(ctx, labID)
	if err != nil {
		return err
	}

	material, err := lm.materialRepo.GetByID(ctx, materialID)
	if err != nil {
		return err
	}
	if lab.CourseID != material.CourseID {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Material does not belong to this lab's course",
		})
	}
	return nil
}

func (lm *labMaterialService) Create(ctx context.Context, req *requests.SetLabMaterial, userID string, labID string) error {
	err := lm.rowExists(ctx, labID, req.MaterialID)
	if err != nil {
		return err
	}

	err = lm.mutationPermission(ctx, userID, labID)
	if err != nil {
		return err
	}

	labMaterial, err := lm.labMaterialRepo.GetByID(ctx, labID, req.MaterialID)
	if err != nil && labMaterial != nil {
		return err
	}
	if labMaterial != nil {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusConflict,
			Message:    "Item already exists",
		})
	}

	ID, err := uuid.NewV7()
	if err != nil {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Cannot generate uuid",
		})
	}

	maxPos, err := lm.labMaterialRepo.MaxPositionByLabID(ctx, labID)
	if err != nil {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Cannot determine material position",
		})
	}

	return lm.labMaterialRepo.Create(ctx, req, ID.String(), labID, maxPos+1)
}

func (lm *labMaterialService) UpdatePosition(ctx context.Context, labID string, materialID string, userID string, newPos int) error {
	err := lm.mutationPermission(ctx, userID, labID)
	if err != nil {
		return err
	}

	labMaterial, err := lm.labMaterialRepo.GetByID(ctx, labID, materialID)
	if err != nil {
		return err
	}

	oldPos := labMaterial.Position
	if newPos == oldPos {
		return nil
	}

	err = lm.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		maxPos, err := lm.labMaterialRepo.MaxPositionByLabID(ctx, labID)
		if err != nil {
			return err
		}

		if newPos >= maxPos {
			newPos = maxPos
		}
		if newPos < 1 {
			newPos = 1
		}

		if newPos < oldPos {
			err = lm.labMaterialRepo.ShiftRangeDown(ctx, labID, newPos, oldPos-1)
		} else {
			err = lm.labMaterialRepo.ShiftRangeUp(ctx, labID, oldPos+1, newPos)
		}
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return lm.labMaterialRepo.UpdatePosition(ctx, labMaterial.ID, newPos)
}

func (lm *labMaterialService) GetByLabAndMaterialID(ctx context.Context, labID string, materialID string) (*models.LabMaterial, error) {
	err := lm.rowExists(ctx, labID, materialID)
	if err != nil {
		return nil, err
	}
	labMat, err := lm.labMaterialRepo.GetByID(ctx, labID, materialID)
	if err != nil {
		return nil, err
	}
	return labMat, nil
}

func (lm *labMaterialService) Delete(ctx context.Context, labID string, userID string, req *requests.DeleteLabMaterial) error {
	err := lm.rowExists(ctx, labID, req.MaterialID)
	if err != nil {
		return err
	}

	err = lm.mutationPermission(ctx, userID, labID)
	if err != nil {
		return err
	}

	labMaterial, err := lm.labMaterialRepo.GetByID(ctx, labID, req.MaterialID)
	if err != nil {
		return err
	}

	err = lm.labMaterialRepo.DeleteByID(ctx, labMaterial.ID)
	if err != nil {
		return err
	}
	return nil
}

func (lm *labMaterialService) GetByLabID(ctx context.Context, labID string) ([]models.LabMaterial, error) {
	_, err := lm.labRepo.GetByID(ctx, labID)
	if err != nil {
		return nil, err
	}
	labMats, err := lm.labMaterialRepo.GetByLabID(ctx, labID)
	if err != nil {
		return nil, err
	}

	for i := range labMats {
		tags, err := lm.readMaterialTagRepo.GetTags(ctx, labMats[i].MaterialID)
		if err != nil {
			return nil, err
		}
		if tags == nil {
			tags = []string{}
		}
		labMats[i].MaterialData.Tags = tags

		mat, err := lm.materialRepo.GetByID(ctx, labMats[i].MaterialID)
		if err != nil {
			return nil, err
		}
		creator, err := lm.userRepo.GetByID(ctx, mat.CreatedBy)
		if err != nil {
			return nil, err
		}
		labMats[i].MaterialData.CreatedBy = &models.MaterialCreator{
			ID:           creator.ID,
			DisplayName:  creator.DisplayName,
			ProfileImage: creator.ProfileImage,
		}
	}

	return labMats, nil
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

	for i := range labMaterials {
		mat, err := lm.materialRepo.GetByID(ctx, labMaterials[i].MaterialID)
		if err != nil {
			return nil, err
		}

		matTags, err := lm.readMaterialTagRepo.GetTags(ctx, mat.ID)
		if err != nil {
			return nil, err
		}

		var tags []string = matTags
		if tags == nil {
			tags = []string{}
		}
		matJson := &models.Material{
			ID:                   mat.ID,
			CourseID:             mat.CourseID,
			ForkedFromMaterialID: mat.ForkedFromMaterialID,
			Name:                 mat.Name,
			Tags:                 tags,
			Type:                 mat.Type,
			Visibility:           mat.Visibility,
			CreatedAt:            mat.CreatedAt,
		}
		labMaterials[i].MaterialData = matJson
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
