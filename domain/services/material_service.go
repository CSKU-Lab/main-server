package services

import (
	"context"
	"net/http"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/google/uuid"
)

type MaterialService interface {
	Create(ctx context.Context, createdByUserID string, req *requests.CreateMaterial) (string, error)
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.Material, error)
	Count(ctx context.Context, search string, filters map[string]string) (int, error)
	GetByID(ctx context.Context, ID string) (*models.MaterialDetail, error)
	UpdateByID(ctx context.Context, ID string, req *requests.BaseUpdateMaterial, rawReq []byte, userID string) error
	DeleteByID(ctx context.Context, ID string, userID string) error
}

type materialService struct {
	repo                repositories.MaterialRepository
	uowRepo             repositories.UoWRepository
	readMaterialTagRepo repositories.ReadMaterialTagRepository
	userRepo            repositories.User
	materialRegistry    registries.Material
	allowedFilterFields map[string]bool
}

func NewMaterialService(repo repositories.MaterialRepository, readMaterialTagRepo repositories.ReadMaterialTagRepository, uowRepo repositories.UoWRepository, userRepo repositories.User, materialRegistry registries.Material) MaterialService {
	return &materialService{
		repo:                repo,
		uowRepo:             uowRepo,
		readMaterialTagRepo: readMaterialTagRepo,
		userRepo:            userRepo,
		materialRegistry:    materialRegistry,
		allowedFilterFields: map[string]bool{
			"name": true,
			"type": true,
		},
	}
}

func (s *materialService) Create(ctx context.Context, createdByUserID string, req *requests.CreateMaterial) (string, error) {
	// materialHandler, exists := s.materialRegistry.GetHandler(req.Type)
	// if !exists {
	// 	return "", cserrors.New(&cserrors.Option{
	// 		HttpStatus: http.StatusBadRequest,
	// 		Message:    "Unsupported material type",
	// 	})
	// }

	var matID string
	err := s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		id, err := uuid.NewV7()
		if err != nil {
			return cserrors.New(&cserrors.Option{
				Message:    "Failed to generate UUID",
				HttpStatus: http.StatusInternalServerError,
			})
		}

		err = u.Material().Create(ctx, id.String(), createdByUserID, req)
		if err != nil {
			return err
		}
		matID = id.String()

		err = u.MaterialTag().SetTags(ctx, id.String(), req.Tags)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	// I think there is other way to do this with more conistency. Stay tuned !
	// err = materialHandler.Create(ctx, matID, req, nil)
	// if err != nil {
	// 	return "", err
	// }

	return matID, nil
}

func (s *materialService) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.Material, error) {
	allowedSortFields := map[string]bool{
		"name":         true,
		"type":         true,
		"started_date": true,
		"created_at":   true,
		"updated_at":   true,
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

	materials, err := s.repo.GetPagination(ctx, page, limit, search, sanitizedSortBy, sanitizedSortOrder, filters)
	if err != nil {
		return nil, err
	}

	matModels := make([]models.Material, 0, len(materials))

	for _, mat := range materials {
		matTags, err := s.readMaterialTagRepo.GetTags(ctx, mat.ID)
		if err != nil {
			return nil, err
		}

		var tags []string = matTags
		if tags == nil {
			tags = []string{}
		}

		creator, err := s.userRepo.GetByID(ctx, mat.CreatedBy)
		if err != nil {
			return nil, err
		}

		matModels = append(matModels, models.Material{
			ID:         mat.ID,
			Name:       mat.Name,
			Tags:       tags,
			Type:       mat.Type,
			Visibility: mat.Visibility,
			CreatedAt:  mat.CreatedAt,
			CreatedBy:  creator.DisplayName,
		})
	}

	return matModels, nil
}

func (s *materialService) Count(ctx context.Context, search string, filterParams map[string]string) (int, error) {
	filters, err := sanitize.Filters(filterParams, s.allowedFilterFields)
	if err != nil {
		return 0, err
	}

	return s.repo.Count(ctx, search, filters)
}

func (s *materialService) GetByID(ctx context.Context, ID string) (*models.MaterialDetail, error) {
	mat, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	materialHandler, exists := s.materialRegistry.GetHandler(mat.Type)
	if !exists {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Unsupported material type",
		})
	}

	creator, err := s.userRepo.GetByID(ctx, mat.CreatedBy)
	if err != nil {
		return nil, err
	}

	matTags, err := s.readMaterialTagRepo.GetTags(ctx, mat.ID)
	if err != nil {
		return nil, err
	}

	var tags []string = matTags
	if tags == nil {
		tags = []string{}
	}

	matModel := &models.MaterialDetail{
		Material: &models.Material{
			ID:         mat.ID,
			Name:       mat.Name,
			Type:       mat.Type,
			Tags:       tags,
			Visibility: mat.Visibility,
			CreatedAt:  mat.CreatedAt,
			CreatedBy:  creator.DisplayName,
		},
		Payload: nil,
	}

	res, err := materialHandler.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	matModel.Payload = res

	return matModel, nil
}

func isUpdateBaseMaterial(req *requests.BaseUpdateMaterial) bool {
	return req.Name != "" || req.Tags != nil || req.Visibility != ""
}

func (s *materialService) UpdateByID(ctx context.Context, ID string, req *requests.BaseUpdateMaterial, rawReq []byte, userID string) error {
	mat, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return err
	}
	if mat.CreatedBy != userID {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusUnauthorized,
			Message:    "No Permission",
		})
	}

	if isUpdateBaseMaterial(req) {
		err := s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
			err := u.Material().UpdateByID(ctx, ID, req)
			if err != nil {
				return err
			}

			if req.Tags != nil {
				return u.MaterialTag().SetTags(ctx, ID, *req.Tags)
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	materialHandler, exists := s.materialRegistry.GetHandler(mat.Type)
	if !exists {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Unsupported material type",
		})
	}

	return materialHandler.UpdateByID(ctx, ID, req, rawReq)
}

func (s *materialService) DeleteByID(ctx context.Context, ID string, userID string) error {
	mat, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return err
	}
	if mat.CreatedBy != userID {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusUnauthorized,
			Message:    "No Permission",
		})
	}
	return s.repo.DeleteByID(ctx, ID)
}
