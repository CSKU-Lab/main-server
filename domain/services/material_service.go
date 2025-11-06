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

type MaterialService interface {
	Create(ctx context.Context, createdByUserID string, req *requests.CreateMaterial) (string, error)
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.Material, error)
	Count(ctx context.Context, search string, filters map[string]string) (int, error)
	GetByID(ctx context.Context, ID string) (*models.Material, error)
	UpdateByID(ctx context.Context, ID string, req *requests.UpdateMaterial) error
	DeleteByID(ctx context.Context, ID string) error
}

type materialService struct {
	repo                repositories.MaterialRepository
	uowRepo             repositories.UoWRepository
	readMaterialTagRepo repositories.ReadMaterialTagRepository
	userRepo            repositories.User
	allowedFilterFields map[string]bool
}

func NewMaterialService(repo repositories.MaterialRepository, readMaterialTagRepo repositories.ReadMaterialTagRepository, uowRepo repositories.UoWRepository, userRepo repositories.User) MaterialService {
	return &materialService{
		repo:                repo,
		uowRepo:             uowRepo,
		readMaterialTagRepo: readMaterialTagRepo,
		userRepo:            userRepo,
		allowedFilterFields: map[string]bool{
			"name": true,
			"type": true,
		},
	}
}

func (s *materialService) Create(ctx context.Context, createdByUserID string, req *requests.CreateMaterial) (string, error) {
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

		return u.MaterialTag().SetTags(ctx, id.String(), req.Tags)
	})
	if err != nil {
		return "", err
	}

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

		tags := []string{}
		if matTags != nil {
			tags = matTags
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

func (s *materialService) GetByID(ctx context.Context, ID string) (*models.Material, error) {
	mat, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	creator, err := s.userRepo.GetByID(ctx, mat.CreatedBy)
	if err != nil {
		return nil, err
	}

	matModel := &models.Material{
		ID:         mat.ID,
		Name:       mat.Name,
		Type:       mat.Type,
		Visibility: mat.Visibility,
		CreatedAt:  mat.CreatedAt,
		CreatedBy:  creator.DisplayName,
	}

	matTags, err := s.readMaterialTagRepo.GetTags(ctx, ID)
	if err != nil {
		return nil, err
	}

	matModel.Tags = matTags

	return matModel, nil
}

func (s *materialService) UpdateByID(ctx context.Context, ID string, req *requests.UpdateMaterial) error {
	return s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		err := u.Material().UpdateByID(ctx, ID, req)
		if err != nil {
			return err
		}

		return u.MaterialTag().SetTags(ctx, ID, req.Tags)
	})
}

func (s *materialService) DeleteByID(ctx context.Context, ID string) error {
	return s.repo.DeleteByID(ctx, ID)
}
