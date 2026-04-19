package services

import (
	"context"
	"net/http"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/registrables"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/google/uuid"
)

type MaterialService interface {
	Create(ctx context.Context, createdByUserID string, req *requests.CreateMaterial) (string, error)
	GetPagination(ctx context.Context, viewerID string, viewerRoles []models.Role, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.Material, error)
	Count(ctx context.Context, viewerID string, viewerRoles []models.Role, search string, filters map[string]string) (int, error)
	GetByID(ctx context.Context, ID string) (*models.MaterialDetail, error)
	GetMaterialWithLatestSubmissionStatus(ctx context.Context, userID string, materialID string, labID string, sectionID string) (*models.MaterialWithSubmissionStatus, error)
	UpdateByID(ctx context.Context, ID string, req *requests.BaseUpdateMaterial, rawReq []byte, userID string) error
	DeleteByID(ctx context.Context, ID string, userID string) error
}

type materialService struct {
	repo                repositories.MaterialRepository
	submissionRepo      repositories.SubmissionRepository
	uowRepo             repositories.UoWRepository
	readMaterialTagRepo repositories.ReadMaterialTagRepository
	userRepo            repositories.User
	materialRegistry    registries.Material
	allowedFilterFields map[string]bool
}

func NewMaterialService(repo repositories.MaterialRepository, submissionRepo repositories.SubmissionRepository, readMaterialTagRepo repositories.ReadMaterialTagRepository, uowRepo repositories.UoWRepository, userRepo repositories.User, materialRegistry registries.Material) MaterialService {
	return &materialService{
		repo:                repo,
		submissionRepo:      submissionRepo,
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
	materialHandler, exists := s.materialRegistry.GetHandler(req.Type)
	if !exists {
		return "", cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Unsupported material type",
		})
	}

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

	err = materialHandler.Create(ctx, matID, req, nil)
	if err != nil {
		return "", err
	}

	return matID, nil
}

func visibilityFilter(viewerID string, roles []models.Role) *repositories.VisibilityFilter {
	for _, r := range roles {
		if r == models.ADMIN {
			return nil
		}
	}
	for _, r := range roles {
		if r == models.INSTRUCTOR {
			return &repositories.VisibilityFilter{ViewerID: viewerID}
		}
	}
	return &repositories.VisibilityFilter{OnlyPublic: true}
}

func (s *materialService) GetPagination(ctx context.Context, viewerID string, viewerRoles []models.Role, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.Material, error) {
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

	materials, err := s.repo.GetPagination(ctx, page, limit, search, sanitizedSortBy, sanitizedSortOrder, filters, visibilityFilter(viewerID, viewerRoles))
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
			CreatedBy: &models.MaterialCreator{
				ID:           creator.ID,
				DisplayName:  creator.DisplayName,
				ProfileImage: creator.ProfileImage,
			},
			AutoScore:   mat.AutoScore,
			ManualScore: mat.ManualScore,
		})
	}

	return matModels, nil
}

func (s *materialService) Count(ctx context.Context, viewerID string, viewerRoles []models.Role, search string, filterParams map[string]string) (int, error) {
	filters, err := sanitize.Filters(filterParams, s.allowedFilterFields)
	if err != nil {
		return 0, err
	}

	return s.repo.Count(ctx, search, filters, visibilityFilter(viewerID, viewerRoles))
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
			ID:          mat.ID,
			Name:        mat.Name,
			Type:        mat.Type,
			Tags:        tags,
			Visibility:  mat.Visibility,
			AutoScore:   mat.AutoScore,
			ManualScore: mat.ManualScore,
			CreatedAt:   mat.CreatedAt,
			CreatedBy: &models.MaterialCreator{
				ID:           creator.ID,
				DisplayName:  creator.DisplayName,
				ProfileImage: creator.ProfileImage,
			},
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
	return req.Name != "" || req.Tags != nil || req.Visibility != "" || req.AutoScore != nil || req.ManualScore != nil
}

func (s *materialService) UpdateByID(ctx context.Context, ID string, req *requests.BaseUpdateMaterial, rawReq []byte, userID string) error {
	mat, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return err
	}
	if mat.CreatedBy != userID {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Message:    "No Permission",
		})
	}

	materialHandler, exists := s.materialRegistry.GetHandler(mat.Type)
	if !exists {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Unsupported material type",
		})
	}

	// First, update the external task service (this is the source of truth)
	err = materialHandler.UpdateByID(ctx, ID, req, rawReq)
	if err != nil {
		return err
	}

	// Calculate scores from the raw request payload
	scores, err := materialHandler.CalculateScores(rawReq)
	if err != nil {
		return err
	}

	// Set the calculated auto_score on the request
	req.AutoScore = &scores.AutoScore
	// Only set manual_score if it wasn't provided in the request
	if req.ManualScore == nil {
		req.ManualScore = &scores.ManualScore
	}

	// Update the database with the new scores (and other base material fields if provided)
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

	return nil
}

func (s *materialService) DeleteByID(ctx context.Context, ID string, userID string) error {
	mat, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return err
	}
	if mat.CreatedBy != userID {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Message:    "No Permission",
		})
	}
	return s.repo.DeleteByID(ctx, ID)
}

// need to refactor this function, by separate Material from Submissions
func (s *materialService) GetMaterialWithLatestSubmissionStatus(ctx context.Context, userID string, materialID string, labID string, sectionID string) (*models.MaterialWithSubmissionStatus, error) {
	// Get material info (name and type)
	material, err := s.repo.GetByID(ctx, materialID)
	if err != nil {
		return nil, err
	}

	// Get latest submission status
	submissions, err := s.submissionRepo.GetPagination(ctx, userID, materialID, labID, sectionID, 1, 1, "desc")
	if err != nil {
		return nil, err
	}

	status := ""
	if len(submissions) > 0 {
		status = string(submissions[0].Status)
	}

	// Get material payload from registry based on material type
	materialHandler, exists := s.materialRegistry.GetHandler(material.Type)
	if !exists {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Unsupported material type",
		})
	}

	payload, err := materialHandler.GetByID(ctx, materialID)
	if err != nil {
		return nil, err
	}

	// Filter payload to only include allowed fields based on material type
	filteredPayload := s.filterPayloadForUser(material.Type, payload)

	return &models.MaterialWithSubmissionStatus{
		Name:    material.Name,
		Type:    material.Type,
		Status:  status,
		Payload: filteredPayload,
	}, nil
}

func (s *materialService) filterPayloadForUser(materialType string, payload any) any {
	switch materialType {
	case "code":
		if codePayload, ok := payload.(*registrables.CodeMaterialResponse); ok {
			// Expose id, name, and starter files for each allowed runner
			type studentRunner struct {
				ID    string              `json:"id"`
				Name  string              `json:"name"`
				Files []registrables.File `json:"files"`
			}
			filteredRunners := make([]studentRunner, len(codePayload.AllowedRunners))
			for i, runner := range codePayload.AllowedRunners {
				filteredRunners[i] = studentRunner{
					ID:    runner.ID,
					Name:  runner.Name,
					Files: runner.Files,
				}
			}

			return struct {
				Description    *string             `json:"description"`
				AllowedRunners []studentRunner     `json:"allowed_runners"`
				ResourceFiles  []registrables.File `json:"resource_files"`
				Limit          any                 `json:"limit"`
			}{
				Description:    codePayload.Description,
				AllowedRunners: filteredRunners,
				Limit:          codePayload.Limit,
				ResourceFiles:  codePayload.ResourceFiles,
			}
		}
	}

	// For other types or if type assertion fails, return as-is
	return payload
}
