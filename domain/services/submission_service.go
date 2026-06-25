package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	contextkeys "github.com/CSKU-Lab/main-server/context_keys"
	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/adapters/pubsub"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SubmissionService interface {
	Create(ctx context.Context, req *requests.Submission, rawPayload []byte) (string, error)
	Update(ctx context.Context, submissionID string, payload *UpdateSubmissionPayload, rawPayload []byte) error
	Get(ctx context.Context, submissionID string) (*models.Submission, error)
	Listen(ctx context.Context, submissionID string) (<-chan *models.Submission, error)
	GetUserSubmissions(ctx context.Context, userID string, materialID string, labID string, sectionID string, page int, pageSize int, sortOrder string) ([]models.Submission, int, error)
	GetUserSubmissionsWithMaterial(ctx context.Context, userID string, materialID string, labID string, sectionID string, page int, pageSize int, sortOrder string) ([]models.Submission, int, error)
	GetGradebookBySectionID(ctx context.Context, ID string) (*models.Gradebook, error)
	GetLabStudentStatus(ctx context.Context, sectionID, labID string) (*models.LabStudentStatus, error)
	CountCompletedStudentsByLabAndSection(ctx context.Context, labID string, sectionID string) (int, error)
	GetSectionLabMaterialSubmissions(ctx context.Context, sectionID string, labID string, materialID string) ([]models.CMSSectionStudentSubmission, error)
	GetStudentSubmissionsByMaterialSectionLab(ctx context.Context, materialID string, sectionID string, labID string, studentID string, page int, pageSize int, sortBy, sortOrder string) ([]models.StudentSubmission, int, error)
	UpdateManualScore(ctx context.Context, submissionID string, manualScore int) error
	GetMaterialStudentStatus(ctx context.Context, userID, materialID, labID, sectionID string) string
	RegradeByMaterial(ctx context.Context, sectionID, labID, materialID string) error
	DeleteSubmission(ctx context.Context, submissionID string) error
}

type UpdateSubmissionPayload struct {
	Status     models.SubmissionStatus
	MaterialID string
	Payload    any `json:"payload"`
}

type submissionService struct {
	repo                  repositories.SubmissionRepository
	materialRepo          repositories.MaterialRepository
	sectionStudentRepo    repositories.SectionStudentRepository
	uowRepo               repositories.UoWRepository
	registry              registries.SubmissionRegistry
	userRepo              repositories.User
	materialReigstry      registries.Material
	sectionRepo           repositories.SectionRepository
	labSectionRepo        repositories.LabSectionRepository
	labMatRepo            repositories.LabMaterialRepository
	codeSubmissionRepo    repositories.CodeSubmissionRepository
	codeMatRepo           repositories.CodeMaterialRepository
	typingSubmissionRepo  repositories.TypingSubmissionRepository
	ps                    pubsub.PubSub
	systemSettingsService SystemSettingsService
	logger                *zap.SugaredLogger
	allowedSortFields     map[string]bool
}

type SubmissionServiceArgs struct {
	SubmissionRepository       repositories.SubmissionRepository
	MaterialRepository         repositories.MaterialRepository
	UowRepository              repositories.UoWRepository
	SubmissionRegistry         registries.SubmissionRegistry
	SectionStudentRepository   repositories.SectionStudentRepository
	UserRepository             repositories.User
	MaterialRegistry           registries.Material
	SectionRepository          repositories.SectionRepository
	LabSectionRepository       repositories.LabSectionRepository
	LabMaterialRepository      repositories.LabMaterialRepository
	CodeSubmissionRepository   repositories.CodeSubmissionRepository
	CodeMaterialRepository     repositories.CodeMaterialRepository
	TypingSubmissionRepository repositories.TypingSubmissionRepository
	PubSub                     pubsub.PubSub
	SystemSettingsService      SystemSettingsService
	Logger                     *zap.SugaredLogger
}

func NewSubmissionService(args *SubmissionServiceArgs) SubmissionService {
	return &submissionService{
		repo:                  args.SubmissionRepository,
		materialRepo:          args.MaterialRepository,
		uowRepo:               args.UowRepository,
		registry:              args.SubmissionRegistry,
		sectionStudentRepo:    args.SectionStudentRepository,
		userRepo:              args.UserRepository,
		sectionRepo:           args.SectionRepository,
		labSectionRepo:        args.LabSectionRepository,
		labMatRepo:            args.LabMaterialRepository,
		materialReigstry:      args.MaterialRegistry,
		codeSubmissionRepo:    args.CodeSubmissionRepository,
		codeMatRepo:           args.CodeMaterialRepository,
		typingSubmissionRepo:  args.TypingSubmissionRepository,
		ps:                    args.PubSub,
		systemSettingsService: args.SystemSettingsService,
		logger:                args.Logger,
		allowedSortFields: map[string]bool{
			"order":      true,
			"created_at": true,
		},
	}
}

func (s *submissionService) GetGradebookBySectionID(ctx context.Context, ID string) (*models.Gradebook, error) {
	_, err := s.sectionRepo.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	students, err := s.sectionStudentRepo.GetBySectionID(ctx, ID)
	if err != nil {
		return nil, err
	}

	labs, err := s.labSectionRepo.GetBySectionID(ctx, ID)
	if err != nil {
		return nil, err
	}

	res := &models.Gradebook{
		LabCol:     []models.LabCol{},
		StudentRow: []models.StudentRow{},
	}
	labMaterials := make(map[string][]*repositories.Material)

	for _, lab := range labs {
		labMats, err := s.labMatRepo.GetByLabID(ctx, lab.ID)
		if err != nil {
			return nil, err
		}

		var totalMaxAutoScore int
		var totalMaxManualScore int

		for _, lm := range labMats {
			mat, err := s.materialRepo.GetByID(ctx, lm.MaterialID)
			if err != nil {
				return nil, err
			}

			totalMaxAutoScore += mat.AutoScore
			totalMaxManualScore += mat.ManualScore
			labMaterials[lab.ID] = append(labMaterials[lab.ID], mat)
		}

		res.LabCol = append(res.LabCol, models.LabCol{
			LabID:          lab.ID,
			LabName:        lab.DisplayName,
			MaxAutoScore:   totalMaxAutoScore,
			MaxManualScore: totalMaxManualScore,
		})
	}

	for _, student := range students {

		studentRow := models.StudentRow{
			Username:    student.Username,
			DisplayName: student.DisplayName,
		}
		studentRow.LabScores = make(map[string]models.LabScore)

		for _, lab := range labs {

			var totalAutoScore int
			var totalManualScore int

			for _, mat := range labMaterials[lab.ID] {
				submission, err := s.repo.GetLatestOfStudentIDInSectionID(ctx, ID, lab.ID, mat.ID, student.ID)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						submission = nil
					} else {
						return nil, err
					}
				}

				if submission != nil {
					totalManualScore += submission.ManualScore
					totalAutoScore += submission.AutoScore
				}

			}

			studentRow.LabScores[lab.ID] = models.LabScore{
				AutoScore:   totalAutoScore,
				ManualScore: totalManualScore,
			}

		}

		res.StudentRow = append(res.StudentRow, studentRow)
	}

	return res, nil
}

func (s *submissionService) GetLabStudentStatus(ctx context.Context, sectionID, labID string) (*models.LabStudentStatus, error) {
	students, err := s.sectionStudentRepo.GetBySectionID(ctx, sectionID)
	if err != nil {
		return nil, err
	}

	materials, err := s.labMatRepo.GetByLabID(ctx, labID)
	if err != nil {
		return nil, err
	}

	res := &models.LabStudentStatus{
		StudentRows:  make([]models.StudentLabStatus, 0, len(students)),
		MaterialCols: make([]models.MaterialCol, 0, len(materials)),
	}

	for _, mat := range materials {
		res.MaterialCols = append(res.MaterialCols, models.MaterialCol{
			MaterialID:   mat.MaterialID,
			MaterialName: mat.MaterialData.Name,
		})
	}

	for _, student := range students {
		studentRow := models.StudentLabStatus{
			Student: &models.Student{
				ID:           student.ID,
				Username:     student.Username,
				DisplayName:  student.DisplayName,
				ProfileImage: student.ProfileImage,
			},
			MaterialStatuses: make(map[string]models.MaterialStatus),
		}

		for _, mat := range materials {
			submission, err := s.repo.GetLatestOfStudentIDInSectionID(ctx, sectionID, labID, mat.MaterialID, student.ID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					studentRow.MaterialStatuses[mat.MaterialID] = models.MaterialStatus{
						Status:      models.NOT_SUBMITTED,
						SubmittedAt: nil,
					}
				} else {
					return nil, err
				}
			} else {
				studentRow.MaterialStatuses[mat.MaterialID] = models.MaterialStatus{
					Status:      submission.Status,
					SubmittedAt: &submission.CreatedAt,
				}
			}
		}

		res.StudentRows = append(res.StudentRows, studentRow)
	}

	return res, nil
}

func (s *submissionService) Create(ctx context.Context, req *requests.Submission, rawPayload []byte) (string, error) {
	user, ok := ctx.Value(contextkeys.UserKey).(contextkeys.User)
	if !ok {
		return "", errors.New("cannot extract user from context")
	}

	mat, err := s.materialRepo.GetByID(ctx, req.MaterialID)
	if err != nil {
		return "", err
	}

	if req.SectionID != nil {
		_, err := s.sectionStudentRepo.GetBySectionAndStudentID(ctx, *req.SectionID, user.ID)
		if err != nil {
			return "", cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusForbidden,
				Message:    "You do not have access to this section",
			})
		}
	}

	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	payload := &repositories.Submission{
		ID:         id.String(),
		UserID:     user.ID,
		LabID:      req.LabID,
		MaterialID: req.MaterialID,
		SectionID:  req.SectionID,
		CourseID:   req.CourseID,
	}

	err = s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		err = u.Submission().Create(ctx, payload)
		if err != nil {
			return err
		}

		handler, err := s.registry.GetHandler(mat.Type)
		if err != nil {
			return err
		}

		return handler.Create(ctx, u, id.String(), req.MaterialID, rawPayload)
	})
	if err != nil {
		return "", err
	}

	return id.String(), nil
}

// this method doesn't need to check the request because it just use internally
func (s *submissionService) Update(ctx context.Context, submissionID string, payload *UpdateSubmissionPayload, rawPayload []byte) error {
	return s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		err := u.Submission().Update(ctx, &repositories.UpdateSubmissionRequest{
			ID:     submissionID,
			Status: &payload.Status,
		})
		if err != nil {
			return err
		}

		mat, err := s.materialRepo.GetByID(ctx, payload.MaterialID)
		if err != nil {
			return err
		}

		handler, err := s.registry.GetHandler(mat.Type)
		if err != nil {
			return err
		}

		return handler.Update(ctx, u, submissionID, rawPayload)
	})
}

func (s *submissionService) Get(ctx context.Context, submissionID string) (*models.Submission, error) {
	submission, err := s.repo.Get(ctx, submissionID)
	if err != nil {
		return nil, err
	}

	mat, err := s.materialRepo.GetByID(ctx, submission.MaterialID)
	if err != nil {
		return nil, err
	}

	handler, err := s.registry.GetHandler(mat.Type)
	if err != nil {
		return nil, err
	}

	payload, err := handler.Get(ctx, submission.ID, "")
	if err != nil {
		return nil, err
	}

	autoScore := 0
	if mat.Type == "typing" {
		autoScore = submission.AutoScore
	}

	return &models.Submission{
		ID:        submission.ID,
		Status:    submission.Status,
		Order:     submission.Order,
		AutoScore: autoScore,
		CreatedAt: submission.CreatedAt,
		Payload:   payload,
	}, nil
}

func (s *submissionService) Listen(ctx context.Context, submissionID string) (<-chan *models.Submission, error) {
	errChan := make(chan error, 1)
	subChan := make(chan *models.Submission)

	repoSubmission, err := s.repo.Get(ctx, submissionID)
	if err != nil {
		return nil, err
	}

	// Get material to determine handler for extracting stats
	mat, err := s.materialRepo.GetByID(ctx, repoSubmission.MaterialID)
	if err != nil {
		errChan <- err
		return nil, err
	}

	handler, err := s.registry.GetHandler(mat.Type)
	if err != nil {
		errChan <- err
		return nil, err
	}

	if repoSubmission.Status == models.PASSED || repoSubmission.Status == models.FAILED {
		go func() {
			subChan <- &models.Submission{
				ID:        submissionID,
				Status:    repoSubmission.Status,
				CreatedAt: repoSubmission.CreatedAt,
			}
			close(subChan)
		}()

		return subChan, nil
	}

	go func() {
		defer close(subChan)

		channel := fmt.Sprintf("submissions:update:%s", submissionID)
		msgs, err := s.ps.Subscribe(ctx, channel)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			errChan <- err
		}

		for msg := range msgs {
			stats := handler.GetOverviewStatsByID(ctx, submissionID)
			status := models.SubmissionStatus(string(msg))

			response := &models.Submission{
				ID:        submissionID,
				Status:    status,
				CreatedAt: repoSubmission.CreatedAt,
				Payload:   stats,
			}

			subChan <- response

			if status == models.FAILED || status == models.PASSED {
				return
			}
		}
	}()

	return subChan, nil
}

func (s *submissionService) GetUserSubmissions(ctx context.Context, userID string, materialID string, labID string, sectionID string, page int, pageSize int, sortOrder string) ([]models.Submission, int, error) {
	submissions, err := s.repo.GetPagination(ctx, userID, materialID, labID, sectionID, page, pageSize, sortOrder)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.repo.Count(ctx, userID, materialID, labID, sectionID)
	if err != nil {
		return nil, 0, err
	}

	if len(submissions) == 0 {
		return []models.Submission{}, count, nil
	}

	var matType string
	if materialID != "" {
		mat, err := s.materialRepo.GetByID(ctx, materialID)
		if err == nil {
			matType = mat.Type
		}
	}

	result := make([]models.Submission, len(submissions))
	for i, sub := range submissions {
		autoScore := 0
		if matType == "typing" {
			autoScore = sub.AutoScore
		}
		result[i] = models.Submission{
			ID:        sub.ID,
			Status:    sub.Status,
			Order:     sub.Order,
			AutoScore: autoScore,
			CreatedAt: sub.CreatedAt,
		}
	}

	return result, count, nil
}

func (s *submissionService) GetUserSubmissionsWithMaterial(ctx context.Context, userID string, materialID string, labID string, sectionID string, page int, pageSize int, sortOrder string) ([]models.Submission, int, error) {
	mat, err := s.materialRepo.GetByID(ctx, materialID)
	if err != nil {
		return nil, 0, err
	}

	submissions, err := s.repo.GetPagination(ctx, userID, materialID, labID, sectionID, page, pageSize, sortOrder)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.repo.Count(ctx, userID, materialID, labID, sectionID)
	if err != nil {
		return nil, 0, err
	}

	if len(submissions) == 0 {
		return []models.Submission{}, count, nil
	}

	handler, err := s.registry.GetHandler(mat.Type)
	if err != nil {
		return nil, 0, err
	}

	submissionIDs := make([]string, len(submissions))
	for i, sub := range submissions {
		submissionIDs[i] = sub.ID
	}

	payloads, err := handler.GetOverviewsPayload(ctx, submissionIDs)
	if err != nil {
		return nil, 0, err
	}

	result := make([]models.Submission, len(submissions))
	for i, sub := range submissions {
		autoScore := 0
		if mat.Type == "typing" {
			autoScore = sub.AutoScore
		}
		result[i] = models.Submission{
			ID:        sub.ID,
			Status:    sub.Status,
			Order:     sub.Order,
			AutoScore: autoScore,
			CreatedAt: sub.CreatedAt,
			Payload:   payloads[sub.ID],
		}
	}

	return result, count, nil
}

func (s *submissionService) GetMaterialStudentStatus(ctx context.Context, userID, materialID, labID, sectionID string) string {
	subs, _, err := s.GetUserSubmissions(ctx, userID, materialID, labID, sectionID, 1, 1, "desc")
	if err != nil || len(subs) == 0 {
		return "not_started"
	}
	if subs[0].Status == models.PASSED {
		return "passed"
	}
	return "not_passed"
}

func (s *submissionService) CountCompletedStudentsByLabAndSection(ctx context.Context, labID string, sectionID string) (int, error) {
	return s.repo.CountCompletedStudentsByLabAndSection(ctx, labID, sectionID)
}

func (s *submissionService) UpdateManualScore(ctx context.Context, submissionID string, manualScore int) error {
	// First, get the submission to find the material ID
	submission, err := s.repo.Get(ctx, submissionID)
	if err != nil {
		return err
	}

	// Get the material to check the maximum manual score
	material, err := s.materialRepo.GetByID(ctx, submission.MaterialID)
	if err != nil {
		return err
	}

	// Validate that the manual score does not exceed the material's maximum
	if manualScore > material.ManualScore {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    fmt.Sprintf("Manual score cannot exceed maximum of %d", material.ManualScore),
		})
	}

	return s.repo.Update(ctx, &repositories.UpdateSubmissionRequest{
		ID:          submissionID,
		ManualScore: &manualScore,
	})
}

func (s *submissionService) GetSectionLabMaterialSubmissions(ctx context.Context, sectionID string, labID string, materialID string) ([]models.CMSSectionStudentSubmission, error) {
	_, err := s.labMatRepo.GetByID(ctx, labID, materialID)
	if err != nil {
		return nil, err
	}

	students, err := s.sectionStudentRepo.GetBySectionID(ctx, sectionID)
	if err != nil {
		return nil, err
	}

	mat, err := s.materialRepo.GetByID(ctx, materialID)
	if err != nil {
		return nil, err
	}

	handler, err := s.registry.GetHandler(mat.Type)
	if err != nil {
		return nil, err
	}

	var rawSubmissions []models.RawSubmission

	// For typing materials, get best submissions; for others, get latest
	if mat.Type == "typing" {
		bestSubs, err := s.typingSubmissionRepo.GetBestByMaterial(ctx, materialID, labID, sectionID)
		if err != nil {
			return nil, err
		}

		// Convert map data to RawSubmission structs
		for _, best := range bestSubs {
			rawSubmissions = append(rawSubmissions, models.RawSubmission{
				ID:          best["id"].(string),
				UserID:      best["user_id"].(string),
				MaterialID:  best["material_id"].(string),
				LabID:       best["lab_id"].(string),
				SectionID:   best["section_id"].(*string),
				CourseID:    best["course_id"].(*string),
				Status:      models.SubmissionStatus(best["status"].(string)),
				Order:       best["submission_order"].(int),
				CreatedAt:   best["created_at"].(time.Time),
				UpdatedAt:   best["updated_at"].(time.Time),
				IPAddress:   best["ip_address"].(string),
				ManualScore: best["manual_score"].(int),
				AutoScore:   best["auto_score"].(int),
			})
		}
	} else {
		var err error
		rawSubmissions, err = s.repo.GetLatestByMaterialSectionAndLabID(ctx, materialID, sectionID, labID)
		if err != nil {
			return nil, err
		}
	}

	submissionMap := make(map[string]models.RawSubmission)
	for _, sub := range rawSubmissions {
		submissionMap[sub.UserID] = sub
	}

	result := make([]models.CMSSectionStudentSubmission, len(students))
	for i, student := range students {
		sub, hasSubmission := submissionMap[student.ID]

		if hasSubmission {
			payload, err := handler.Get(ctx, sub.ID, "instructor")
			if err != nil {
				return nil, err
			}

			result[i] = models.CMSSectionStudentSubmission{
				Student: student,
				StudentSubmission: &models.StudentSubmission{
					Order:       sub.Order,
					AutoScore:   sub.AutoScore,
					ManualScore: sub.ManualScore,
					IP:          sub.IPAddress,
					Status:      sub.Status,
					CreatedAt:   sub.CreatedAt,
					UpdatedAt:   sub.UpdatedAt,
					Payload:     payload,
				},
			}
		} else {
			result[i] = models.CMSSectionStudentSubmission{
				Student: student,
				StudentSubmission: &models.StudentSubmission{
					AutoScore:   0,
					ManualScore: 0,
					IP:          "",
					Status:      models.NOT_SUBMITTED,
					Payload:     nil,
				},
			}
		}
	}

	return result, nil
}

func (s *submissionService) GetStudentSubmissionsByMaterialSectionLab(ctx context.Context, materialID string, sectionID string, labID string, studentID string, page int, pageSize int, sortBy, sortOrder string) ([]models.StudentSubmission, int, error) {
	sanitizedSortBy, err := sanitize.SortBy(sortBy, s.allowedSortFields)
	if err != nil {
		return nil, 0, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid sort by field",
			})
	}

	sanitizedSortOrder, err := sanitize.SortOrder(sortOrder)
	if err != nil {
		return nil, 0, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid sort order",
			})
	}

	mat, err := s.materialRepo.GetByID(ctx, materialID)
	if err != nil {
		return nil, 0, err
	}

	rawSubmissions, err := s.repo.GetPaginationByMaterialSectionLabAndStudentID(ctx, materialID, sectionID, labID, studentID, page, pageSize, sanitizedSortBy, sanitizedSortOrder)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.repo.CountByMaterialSectionLabAndStudentID(ctx, materialID, sectionID, labID, studentID)
	if err != nil {
		return nil, 0, err
	}

	if len(rawSubmissions) == 0 {
		return []models.StudentSubmission{}, count, nil
	}

	handler, err := s.registry.GetHandler(mat.Type)
	if err != nil {
		return nil, 0, err
	}

	submissionIDs := make([]string, len(rawSubmissions))
	for i, sub := range rawSubmissions {
		submissionIDs[i] = sub.ID
	}

	payloads, err := handler.GetByIDs(ctx, submissionIDs, "instructor")
	if err != nil {
		return nil, 0, err
	}

	result := make([]models.StudentSubmission, len(rawSubmissions))
	for i, sub := range rawSubmissions {
		result[i] = models.StudentSubmission{
			ID:          sub.ID,
			Status:      sub.Status,
			Order:       sub.Order,
			CreatedAt:   sub.CreatedAt,
			UpdatedAt:   sub.UpdatedAt,
			AutoScore:   sub.AutoScore,
			ManualScore: sub.ManualScore,
			IP:          sub.IPAddress,
			Payload:     payloads[sub.ID],
		}
	}

	return result, count, nil
}

func (s *submissionService) DeleteSubmission(ctx context.Context, submissionID string) error {
	return s.repo.Delete(ctx, submissionID)
}

func (s *submissionService) RegradeByMaterial(ctx context.Context, sectionID, labID, materialID string) error {
	mat, err := s.materialRepo.GetByID(ctx, materialID)
	if err != nil {
		return err
	}
	if mat.Type != "code" {
		return nil
	}

	codeMat, err := s.codeMatRepo.GetByID(ctx, materialID)
	if err != nil {
		return err
	}

	submissions, err := s.repo.GetLatestByMaterialSectionAndLabID(ctx, materialID, sectionID, labID)
	if err != nil {
		return err
	}

	// Detach from the request context: the HTTP handler returns 202 immediately,
	// after which Fiber recycles the underlying *fasthttp.RequestCtx. These
	// goroutines outlive the request, so using the request ctx would make every
	// DB op run against a cancelled/recycled context and silently fail.
	// WithoutCancel keeps request values (tracing, etc.) but drops the deadline.
	bgCtx := context.WithoutCancel(ctx)

	for _, sub := range submissions {
		sub := sub
		go func() {
			codeSub, err := s.codeSubmissionRepo.Get(bgCtx, sub.ID)
			if err != nil {
				s.logger.Errorw("regrade: failed to load code submission",
					"submissionID", sub.ID, "materialID", materialID, "error", err)
				return
			}
			if codeSub.RunnerID == nil || *codeSub.RunnerID == "" {
				return
			}

			id, err := uuid.NewV7()
			if err != nil {
				s.logger.Errorw("regrade: failed to generate id",
					"submissionID", sub.ID, "materialID", materialID, "error", err)
				return
			}

			queued := models.QUEUED
			gradePayload := &models.GradeExecution{
				ID:              id.String(),
				Files:           codeSub.Files,
				RunnerID:        *codeSub.RunnerID,
				TaskID:          codeMat.TaskID,
				CompareScriptID: s.systemSettingsService.GetDefaultCompareScriptID(bgCtx),
			}

			if err := s.uowRepo.Execute(bgCtx, func(u repositories.UoWInstance) error {
				if err := u.Submission().Update(bgCtx, &repositories.UpdateSubmissionRequest{
					ID:     sub.ID,
					Status: &queued,
				}); err != nil {
					return err
				}
				return u.CodeSubmissionOutbox().Create(bgCtx, id.String(), sub.ID, gradePayload)
			}); err != nil {
				s.logger.Errorw("regrade: failed to queue submission",
					"submissionID", sub.ID, "materialID", materialID, "error", err)
			}
		}()
	}

	return nil
}
