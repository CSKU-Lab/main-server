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
	"github.com/CSKU-Lab/queue"
	"github.com/google/uuid"
)

type MaterialService interface {
	Create(ctx context.Context, courseID string, createdByUserID string, req *requests.CreateMaterial) (string, error)
	Fork(ctx context.Context, targetCourseID string, sourceMaterialID string, user *models.User) (string, error)
	Clone(ctx context.Context, courseID string, sourceMaterialID string, user *models.User) (string, error)
	GetPagination(ctx context.Context, courseID string, viewerID string, viewerRoles []models.Role, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.Material, error)
	Count(ctx context.Context, courseID string, viewerID string, viewerRoles []models.Role, search string, filters map[string]string) (int, error)
	GetByID(ctx context.Context, courseID string, ID string) (*models.MaterialDetail, error)
	GetByIDUnscoped(ctx context.Context, ID string) (*models.MaterialDetail, error)
	GetMaterialWithLatestSubmissionStatus(ctx context.Context, userID string, materialID string, labID string, sectionID string) (*models.MaterialWithSubmissionStatus, error)
	UpdateByID(ctx context.Context, courseID string, ID string, req *requests.BaseUpdateMaterial, rawReq []byte, userID string) error
	DeleteByID(ctx context.Context, courseID string, ID string, userID string) error
}

type materialService struct {
	repo                repositories.MaterialRepository
	submissionRepo      repositories.SubmissionRepository
	uowRepo             repositories.UoWRepository
	readMaterialTagRepo repositories.ReadMaterialTagRepository
	userRepo            repositories.User
	materialRegistry    registries.Material
	allowedFilterFields map[string]bool
	q                   queue.Queue
}

func NewMaterialService(repo repositories.MaterialRepository, submissionRepo repositories.SubmissionRepository, readMaterialTagRepo repositories.ReadMaterialTagRepository, uowRepo repositories.UoWRepository, userRepo repositories.User, materialRegistry registries.Material, q queue.Queue) MaterialService {
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
		q: q,
	}
}

func (s *materialService) Create(ctx context.Context, courseID string, createdByUserID string, req *requests.CreateMaterial) (string, error) {
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

		err = u.Material().Create(ctx, id.String(), courseID, createdByUserID, nil, req)
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
		// Best-effort cleanup: soft-delete the base material row to prevent orphan
		_ = s.repo.DeleteByID(ctx, matID)
		return "", err
	}

	tag := "Lab Material"
	publishOGEvent(s.q, ogImageEvent{Type: "material", ID: matID, Title: req.Name, Tag: &tag})

	return matID, nil
}

// Fork copies a material into targetCourseID, preserving its name. Used to
// reuse a material across courses; lineage is tracked via forked_from_material_id.
func (s *materialService) Fork(ctx context.Context, targetCourseID string, sourceMaterialID string, user *models.User) (string, error) {
	return s.duplicate(ctx, targetCourseID, sourceMaterialID, user, "")
}

// Clone duplicates a material within the same course, appending " (Clone)" to
// the name so the copy is distinguishable from the source. Copies all config,
// including code config, via the type-specific Clone handler.
func (s *materialService) Clone(ctx context.Context, courseID string, sourceMaterialID string, user *models.User) (string, error) {
	return s.duplicate(ctx, courseID, sourceMaterialID, user, " (Clone)")
}

// duplicate creates a copy of sourceMaterialID in targetCourseID. nameSuffix is
// appended to the source name (empty for a plain fork).
func (s *materialService) duplicate(ctx context.Context, targetCourseID string, sourceMaterialID string, user *models.User, nameSuffix string) (string, error) {
	source, err := s.repo.GetByID(ctx, sourceMaterialID)
	if err != nil {
		return "", err
	}

	if err := s.canViewMaterial(ctx, source, user); err != nil {
		return "", err
	}

	tags, err := s.readMaterialTagRepo.GetTags(ctx, source.ID)
	if err != nil {
		return "", err
	}
	if tags == nil {
		tags = []string{}
	}

	req := &requests.CreateMaterial{
		Name:        source.Name + nameSuffix,
		Tags:        tags,
		Type:        source.Type,
		Visibility:  source.Visibility,
		ManualScore: source.ManualScore,
	}

	materialHandler, exists := s.materialRegistry.GetHandler(source.Type)
	if !exists {
		return "", cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Unsupported material type",
		})
	}

	var targetID string
	err = s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		id, err := uuid.NewV7()
		if err != nil {
			return cserrors.New(&cserrors.Option{
				Message:    "Failed to generate UUID",
				HttpStatus: http.StatusInternalServerError,
			})
		}
		targetID = id.String()

		if err := u.Material().Create(ctx, targetID, targetCourseID, user.ID, &source.ID, req); err != nil {
			return err
		}

		if err := u.MaterialTag().SetTags(ctx, targetID, tags); err != nil {
			return err
		}

		updateReq := &requests.BaseUpdateMaterial{
			AutoScore:   &source.AutoScore,
			ManualScore: &source.ManualScore,
		}
		return u.Material().UpdateByID(ctx, targetID, updateReq)
	})
	if err != nil {
		return "", err
	}

	if err := materialHandler.Clone(ctx, source.ID, targetID); err != nil {
		_ = s.repo.DeleteByID(ctx, targetID)
		return "", err
	}

	return targetID, nil
}

func (s *materialService) canViewMaterial(ctx context.Context, material *repositories.Material, user *models.User) error {
	for _, role := range user.Roles {
		if role == models.ADMIN {
			return nil
		}
	}

	if material.Visibility == "public" || material.CreatedBy == user.ID {
		return nil
	}

	creators, err := s.uowCourseCreators(ctx, material.CourseID)
	if err != nil {
		return err
	}
	for _, creator := range creators {
		if creator.ID == user.ID {
			return nil
		}
	}

	return cserrors.New(&cserrors.Option{
		HttpStatus: http.StatusForbidden,
		Message:    "No Permission",
	})
}

func (s *materialService) uowCourseCreators(ctx context.Context, courseID string) ([]models.CourseCreator, error) {
	var creators []models.CourseCreator
	err := s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		var err error
		creators, err = u.CourseCreator().GetCreators(ctx, courseID)
		return err
	})
	return creators, err
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

func (s *materialService) GetPagination(ctx context.Context, courseID string, viewerID string, viewerRoles []models.Role, page int, limit int, search string, sortBy string, sortOrder string, filterParams map[string]string) ([]models.Material, error) {
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
		if t != "document" && t != "code" && t != "typing" {
			return nil, cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid material type filter",
			})
		}
		filterParams["type__is"] = t
	}

	filters, err := sanitize.Filters(filterParams, s.allowedFilterFields)
	if err != nil {
		return nil, err
	}

	materials, err := s.repo.GetPagination(ctx, courseID, page, limit, search, sanitizedSortBy, sanitizedSortOrder, filters, visibilityFilter(viewerID, viewerRoles))
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
			ID:                   mat.ID,
			CourseID:             mat.CourseID,
			ForkedFromMaterialID: mat.ForkedFromMaterialID,
			Name:                 mat.Name,
			Tags:                 tags,
			Type:                 mat.Type,
			Visibility:           mat.Visibility,
			CreatedAt:            mat.CreatedAt,
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

func (s *materialService) Count(ctx context.Context, courseID string, viewerID string, viewerRoles []models.Role, search string, filterParams map[string]string) (int, error) {
	filters, err := sanitize.Filters(filterParams, s.allowedFilterFields)
	if err != nil {
		return 0, err
	}

	return s.repo.Count(ctx, courseID, search, filters, visibilityFilter(viewerID, viewerRoles))
}

func (s *materialService) GetByID(ctx context.Context, courseID string, ID string) (*models.MaterialDetail, error) {
	mat, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	if mat.CourseID != courseID {
		return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusNotFound, Message: "Material not found"})
	}

	return s.materialDetailFromRepo(ctx, mat)
}

func (s *materialService) GetByIDUnscoped(ctx context.Context, ID string) (*models.MaterialDetail, error) {
	mat, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	return s.materialDetailFromRepo(ctx, mat)
}

func (s *materialService) materialDetailFromRepo(ctx context.Context, mat *repositories.Material) (*models.MaterialDetail, error) {
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
			ID:                   mat.ID,
			CourseID:             mat.CourseID,
			ForkedFromMaterialID: mat.ForkedFromMaterialID,
			Name:                 mat.Name,
			Type:                 mat.Type,
			Tags:                 tags,
			Visibility:           mat.Visibility,
			AutoScore:            mat.AutoScore,
			ManualScore:          mat.ManualScore,
			CreatedAt:            mat.CreatedAt,
			CreatedBy: &models.MaterialCreator{
				ID:           creator.ID,
				DisplayName:  creator.DisplayName,
				ProfileImage: creator.ProfileImage,
			},
		},
		Payload: nil,
	}

	res, err := materialHandler.GetByID(ctx, mat.ID)
	if err != nil {
		return nil, err
	}

	matModel.Payload = res

	return matModel, nil
}

func isUpdateBaseMaterial(req *requests.BaseUpdateMaterial) bool {
	return req.Name != "" || req.Tags != nil || req.Visibility != "" || req.AutoScore != nil || req.ManualScore != nil
}

func (s *materialService) UpdateByID(ctx context.Context, courseID string, ID string, req *requests.BaseUpdateMaterial, rawReq []byte, userID string) error {
	mat, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return err
	}
	if mat.CourseID != courseID {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusNotFound, Message: "Material not found"})
	}
	if mat.CreatedBy != userID {
		userData, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return err
		}
		isAdmin := false
		for _, role := range userData.Roles {
			if role == string(models.ADMIN) {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusForbidden,
				Message:    "No Permission",
			})
		}
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

	// Override auto_score when the handler derives it. For code/typing that
	// means score > 0; for document we always propagate (including 0, so that
	// removing all embedded problems resets the score correctly).
	if scores.AutoScore > 0 || mat.Type == "document" {
		req.AutoScore = &scores.AutoScore
	}
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

	name := mat.Name
	if req.Name != "" {
		name = req.Name
	}
	tag := "Lab Material"
	publishOGEvent(s.q, ogImageEvent{Type: "material", ID: ID, Title: name, Tag: &tag})

	return nil
}

func (s *materialService) DeleteByID(ctx context.Context, courseID string, ID string, userID string) error {
	mat, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return err
	}
	if mat.CourseID != courseID {
		return cserrors.New(&cserrors.Option{HttpStatus: http.StatusNotFound, Message: "Material not found"})
	}
	if mat.CreatedBy != userID {
		userData, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return err
		}
		isAdmin := false
		for _, role := range userData.Roles {
			if role == string(models.ADMIN) {
				isAdmin = true
				break
			}
		}
		if !isAdmin {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusForbidden,
				Message:    "No Permission",
			})
		}
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

	// Document materials have no submission of their own — students submit the
	// embedded code blocks individually. Derive the document status by
	// aggregating the latest submission of each embedded code material.
	if material.Type == "document" {
		if doc, ok := payload.(*registrables.DocumentMaterialResponse); ok {
			embedIDs := registrables.ExtractEmbeddedCodeMaterialIDs(doc.Content)
			status = s.aggregateEmbeddedSubmissionStatus(ctx, userID, embedIDs, labID, sectionID)
		}
	}

	// Filter payload to only include allowed fields based on material type
	filteredPayload := s.filterPayloadForUser(material.Type, payload)

	return &models.MaterialWithSubmissionStatus{
		Name:        material.Name,
		Type:        material.Type,
		Status:      status,
		ManualScore: material.ManualScore,
		Payload:     filteredPayload,
	}, nil
}

// aggregateEmbeddedSubmissionStatus rolls the latest submission status of every
// embedded code material into a single status for the parent document. With a
// four-plus-one state model the precedence is:
//
//	any child still grading        -> running   (Grading)
//	every child has passed         -> passed    (Passed)
//	any child failed               -> failed    (Failed)
//	some — but not all — submitted  -> partial   (Partial)
//	no child submitted             -> ""         (No Submission)
//
// Returns "" when the document has no embedded code blocks.
func (s *materialService) aggregateEmbeddedSubmissionStatus(ctx context.Context, userID string, embedIDs []string, labID string, sectionID string) string {
	if len(embedIDs) == 0 {
		return ""
	}

	anyPending := false
	anyFailed := false
	submitted := 0
	passed := 0

	for _, id := range embedIDs {
		subs, err := s.submissionRepo.GetPagination(ctx, userID, id, labID, sectionID, 1, 1, "desc")
		if err != nil || len(subs) == 0 {
			continue
		}
		submitted++
		switch subs[0].Status {
		case models.QUEUED, models.RUNNING:
			anyPending = true
		case models.PASSED:
			passed++
		case models.FAILED:
			anyFailed = true
		}
	}

	switch {
	case anyPending:
		return string(models.RUNNING)
	case passed == len(embedIDs):
		return string(models.PASSED)
	case anyFailed:
		return string(models.FAILED)
	case submitted > 0:
		return string(models.PARTIAL)
	default:
		return ""
	}
}

// sanitizeStudentFiles strips grader-only information from files before they
// are sent to a student.
//
// Segments are never added or removed: the editable-segment indices the client
// submits are positions in this array, and the grader assembles using the same
// positions on the real task segments, so a shift would corrupt assembly. Only
// content and type are adjusted in place:
//
//   - hidden: carries solution/grader code → content blanked. The trailing
//     newline is preserved when present so the client's normalizeHiddenSegments
//     (which folds a following editable's leading "\n" into a hidden segment
//     that lacks one) keeps behaving exactly as it did with the real content,
//     leaving editable indices untouched. Blanked hidden content is excluded
//     from the rebuilt flat File.Content.
//   - exclude: a visible hint that is not compiled. The content stays, but the
//     type is recast to "readonly" so the student cannot tell the code is
//     excluded from grading (which would hint at the solution shape).
func sanitizeStudentFiles(files []registrables.File) []registrables.File {
	sanitized := make([]registrables.File, len(files))
	for i, f := range files {
		if len(f.Segments) == 0 {
			sanitized[i] = f
			continue
		}

		segments := make([]registrables.FileSegment, len(f.Segments))
		var content strings.Builder
		for j, seg := range f.Segments {
			switch seg.Type {
			case "hidden":
				blanked := ""
				if strings.HasSuffix(seg.Content, "\n") {
					blanked = "\n"
				}
				segments[j] = registrables.FileSegment{Content: blanked, Type: seg.Type}
				// excluded from flat content
			case "exclude":
				segments[j] = registrables.FileSegment{Content: seg.Content, Type: "readonly"}
				content.WriteString(seg.Content)
			default:
				segments[j] = seg
				content.WriteString(seg.Content)
			}
		}

		sanitized[i] = registrables.File{
			Name:     f.Name,
			Content:  content.String(),
			Segments: segments,
		}
	}
	return sanitized
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
					Files: sanitizeStudentFiles(runner.Files),
				}
			}

			return struct {
				Description    *string             `json:"description"`
				AllowedRunners []studentRunner     `json:"allowed_runners"`
				ResourceFiles  []registrables.File `json:"resource_files"`
				Limits         any                 `json:"limits"`
			}{
				Description:    codePayload.Description,
				AllowedRunners: filteredRunners,
				Limits:         codePayload.Limits,
				ResourceFiles:  sanitizeStudentFiles(codePayload.ResourceFiles),
			}
		}
	}

	// For other types or if type assertion fails, return as-is
	return payload
}
