package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

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
	GetDocumentAggregateStatus(ctx context.Context, userID, materialID, sectionID, labID string) (string, error)
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
	docMatRepo            repositories.DocumentMaterialRepository
	inputSubmissionRepo   repositories.InputSubmissionRepository
	ps                    pubsub.PubSub
	systemSettingsService SystemSettingsService
	logger                *zap.SugaredLogger
	allowedSortFields     map[string]bool
}

type SubmissionServiceArgs struct {
	SubmissionRepository         repositories.SubmissionRepository
	MaterialRepository           repositories.MaterialRepository
	UowRepository                repositories.UoWRepository
	SubmissionRegistry           registries.SubmissionRegistry
	SectionStudentRepository     repositories.SectionStudentRepository
	UserRepository               repositories.User
	MaterialRegistry             registries.Material
	SectionRepository            repositories.SectionRepository
	LabSectionRepository         repositories.LabSectionRepository
	LabMaterialRepository        repositories.LabMaterialRepository
	CodeSubmissionRepository     repositories.CodeSubmissionRepository
	CodeMaterialRepository       repositories.CodeMaterialRepository
	TypingSubmissionRepository   repositories.TypingSubmissionRepository
	DocumentMaterialRepository   repositories.DocumentMaterialRepository
	InputSubmissionRepository    repositories.InputSubmissionRepository
	PubSub                       pubsub.PubSub
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
		docMatRepo:            args.DocumentMaterialRepository,
		inputSubmissionRepo:   args.InputSubmissionRepository,
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
	// Document materials have no submission of their own — their auto score is
	// the sum of the embedded code materials' latest scores. Collect each
	// document material's embedded code material IDs once per (lab, document)
	// so the student loop can sum their scores from the in-memory index below.
	// docEmbedIDs[labID][documentMaterialID] = embedded code material IDs.
	docEmbedIDs := make(map[string]map[string][]string)

	// Fetch every student's latest submission for the whole section in a single
	// query, then index it in memory. This replaces a per-(student, lab,
	// material) lookup that previously fired thousands of queries per gradebook.
	// latest[labID][materialID][userID] = latest submission.
	allLatest, err := s.repo.GetLatestScoresBySection(ctx, ID)
	if err != nil {
		return nil, err
	}
	latest := make(map[string]map[string]map[string]*models.RawSubmission)
	for i := range allLatest {
		sub := &allLatest[i]
		if latest[sub.LabID] == nil {
			latest[sub.LabID] = make(map[string]map[string]*models.RawSubmission)
		}
		if latest[sub.LabID][sub.MaterialID] == nil {
			latest[sub.LabID][sub.MaterialID] = make(map[string]*models.RawSubmission)
		}
		latest[sub.LabID][sub.MaterialID][sub.UserID] = sub
	}

	// Input embed submissions contribute to a document material's auto score.
	// Fetch every latest input submission for the section once and index the
	// per-student total per (lab, document material).
	// inputScores[labID][documentMaterialID][userID] = sum of input node scores.
	allInput, err := s.inputSubmissionRepo.ListLatestBySection(ctx, ID)
	if err != nil {
		return nil, err
	}
	inputScores := make(map[string]map[string]map[string]int)
	for i := range allInput {
		sub := &allInput[i]
		if inputScores[sub.LabID] == nil {
			inputScores[sub.LabID] = make(map[string]map[string]int)
		}
		if inputScores[sub.LabID][sub.DocumentMaterialID] == nil {
			inputScores[sub.LabID][sub.DocumentMaterialID] = make(map[string]int)
		}
		inputScores[sub.LabID][sub.DocumentMaterialID][sub.UserID] += sub.Score
	}

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

			totalMaxManualScore += mat.ManualScore
			labMaterials[lab.ID] = append(labMaterials[lab.ID], mat)

			if mat.Type == "document" {
				embeds, err := s.docMatRepo.GetEmbeddedMaterialIDs(ctx, mat.ID)
				if err != nil {
					return nil, err
				}
				if len(embeds) > 0 {
					if docEmbedIDs[lab.ID] == nil {
						docEmbedIDs[lab.ID] = make(map[string][]string)
					}
					docEmbedIDs[lab.ID][mat.ID] = embeds
				}

				// A document's max auto score is the sum of its embedded code
				// materials' CURRENT max — the same live source the student
				// score is aggregated from. The document's own stored auto_score
				// is a snapshot taken when the document was last saved; it goes
				// stale when an embedded material's points change afterwards,
				// making the displayed max lower than a student's real score.
				for _, embedID := range embeds {
					embedMat, err := s.materialRepo.GetByID(ctx, embedID)
					if err != nil {
						// Embedded material was deleted but still referenced in
						// the document content — skip it, matching the student
						// aggregation which simply finds no score for it.
						var csErr *cserrors.Error
						if errors.As(err, &csErr) && csErr.HttpStatus == http.StatusNotFound {
							continue
						}
						return nil, err
					}
					totalMaxAutoScore += embedMat.AutoScore
				}

				// Input embed nodes add their configured max score to the document.
				inputNodes, err := s.docInputNodes(ctx, mat.ID)
				if err != nil {
					return nil, err
				}
				for _, n := range inputNodes {
					totalMaxAutoScore += n.Score
				}
				continue
			}

			totalMaxAutoScore += mat.AutoScore
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
				// Document materials: auto score is the sum of embedded code
				// material scores; there is no direct submission to look up.
				if mat.Type == "document" {
					for _, embedID := range docEmbedIDs[lab.ID][mat.ID] {
						if sub := latest[lab.ID][embedID][student.ID]; sub != nil {
							totalAutoScore += sub.AutoScore
						}
					}
					totalAutoScore += inputScores[lab.ID][mat.ID][student.ID]
					continue
				}

				if sub := latest[lab.ID][mat.ID][student.ID]; sub != nil {
					totalManualScore += sub.ManualScore
					totalAutoScore += sub.AutoScore
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

	// Pre-compute embed-derived status maps for document materials so we don't
	// re-query inside the O(students × materials) loop below.
	docEmbedStatusMaps := make(map[string]map[string]models.SubmissionStatus) // matID → userID → status
	for _, mat := range materials {
		if mat.MaterialData != nil && mat.MaterialData.Type == "document" {
			_, statusByUser, _, _ := s.docEmbedScores(ctx, mat.MaterialID, sectionID, labID)
			if statusByUser != nil {
				docEmbedStatusMaps[mat.MaterialID] = statusByUser
			}
		}
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
			// Document materials: derive status from embedded code submissions.
			// The CMS status grid only renders passed/failed/not_submitted, so
			// map PARTIAL → FAILED (some embeds not passing = not fully done).
			if embedStatuses, ok := docEmbedStatusMaps[mat.MaterialID]; ok {
				status, hasStatus := embedStatuses[student.ID]
				if !hasStatus {
					studentRow.MaterialStatuses[mat.MaterialID] = models.MaterialStatus{
						Status:      models.NOT_SUBMITTED,
						SubmittedAt: nil,
					}
				} else {
					mappedStatus := models.FAILED
					if status == models.PASSED {
						mappedStatus = models.PASSED
					}
					studentRow.MaterialStatuses[mat.MaterialID] = models.MaterialStatus{
						Status:      mappedStatus,
						SubmittedAt: nil,
					}
				}
				continue
			}

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
				Payload:   handler.GetOverviewStatsByID(ctx, submissionID),
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
			return
		}

		// Re-check the status now that the subscription is confirmed active. The
		// grader updates the DB before publishing, so if the submission already
		// reached a terminal state, the final publish either landed in the
		// subscribe window (and would otherwise be lost) or is still pending (and
		// will be caught by the loop below). Emitting here closes the
		// subscribe-after-publish race that left the list stuck on "queued".
		latest, err := s.repo.Get(ctx, submissionID)
		if err == nil && (latest.Status == models.PASSED || latest.Status == models.FAILED) {
			subChan <- &models.Submission{
				ID:        submissionID,
				Status:    latest.Status,
				CreatedAt: latest.CreatedAt,
				Payload:   handler.GetOverviewStatsByID(ctx, submissionID),
			}
			return
		}

		for msg := range msgs {
			var event models.SubmissionStatusEvent
			if err := json.Unmarshal(msg, &event); err != nil {
				s.logger.Errorw("Cannot unmarshal submission status event", "error", err, "submission_id", submissionID)
				continue
			}

			response := &models.Submission{
				ID:        submissionID,
				Status:    event.Status,
				CreatedAt: repoSubmission.CreatedAt,
				Payload:   event.Payload,
			}

			subChan <- response

			if event.Status == models.FAILED || event.Status == models.PASSED {
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
	mat, err := s.materialRepo.GetByID(ctx, materialID)
	if err == nil && mat.Type == "document" {
		_, statusByUser, _, _ := s.docEmbedScores(ctx, materialID, sectionID, labID)
		if statusByUser != nil {
			switch statusByUser[userID] {
			case models.PASSED:
				return "passed"
			case models.PARTIAL:
				return "in_progress"
			}
		}
		return "not_started"
	}

	subs, _, err := s.GetUserSubmissions(ctx, userID, materialID, labID, sectionID, 1, 1, "desc")
	if err != nil || len(subs) == 0 {
		return "not_started"
	}
	if subs[0].Status == models.PASSED {
		return "passed"
	}
	return "not_passed"
}

// GetDocumentAggregateStatus returns a document material's status for a single
// user, derived from its embedded code + input submissions — the same source of
// truth the lab grid uses (docEmbedScores). Returns "" (No Submission) when the
// user has submitted nothing. Fixes the doc detail page reporting "No Submission"
// while the grid shows a real status, because the old doc-page path counted only
// code embeds and ignored input embeds.
func (s *submissionService) GetDocumentAggregateStatus(ctx context.Context, userID, materialID, sectionID, labID string) (string, error) {
	_, statusByUser, _, err := s.docEmbedScores(ctx, materialID, sectionID, labID)
	if err != nil {
		return "", err
	}
	if statusByUser == nil {
		return "", nil
	}
	return string(statusByUser[userID]), nil
}

func (s *submissionService) CountCompletedStudentsByLabAndSection(ctx context.Context, labID string, sectionID string) (int, error) {
	// A student "completes" a lab when every material's status is PASSED. This
	// must be derived per-material rather than counting passed `submissions`
	// rows: document materials have no submission row of their own — their
	// status comes from embedded code + input submissions — so a pure-SQL count
	// over the submissions table never reaches the full material count for any
	// lab containing a document, reporting 0 completed. Reuse the same per-student
	// status the CMS grid renders.
	status, err := s.GetLabStudentStatus(ctx, sectionID, labID)
	if err != nil {
		return 0, err
	}
	if len(status.MaterialCols) == 0 {
		return 0, nil
	}

	completed := 0
	for _, row := range status.StudentRows {
		allPassed := true
		for _, col := range status.MaterialCols {
			st, ok := row.MaterialStatuses[col.MaterialID]
			if !ok || st.Status != models.PASSED {
				allPassed = false
				break
			}
		}
		if allPassed {
			completed++
		}
	}
	return completed, nil
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

// docEmbedScores aggregates embedded code submission scores and computes a
// derived status per student for document materials.
// Returns (autoScoreByUser, statusByUser, perEmbedByUser) maps.
// perEmbedByUser[userID][embedMaterialID] = that embed's auto_score for the user.
func (s *submissionService) docEmbedScores(ctx context.Context, materialID, sectionID, labID string) (map[string]int, map[string]models.SubmissionStatus, map[string]map[string]int, error) {
	embeds, err := s.docMatRepo.GetEmbeddedMaterialIDs(ctx, materialID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Input embed nodes live in the document content, not the embedded-code index.
	inputNodes, err := s.docInputNodes(ctx, materialID)
	if err != nil {
		return nil, nil, nil, err
	}

	if len(embeds) == 0 && len(inputNodes) == 0 {
		return nil, nil, nil, nil
	}

	// Per user: sum of auto_scores and aggregate status across all embedded
	// code materials and input embed nodes.
	autoScores := make(map[string]int)
	// passedCount[userID] = number of embeds (code + input) the user has passed.
	passedCount := make(map[string]int)
	// submittedItems[userID] = set of embed keys (materialID or nodeID) submitted.
	submittedItems := make(map[string]map[string]bool)
	// perEmbedByUser[userID][embedMaterialID] = that code embed's latest auto_score.
	// Only code embeds are tracked here; the CMS detail panel keys on material IDs.
	perEmbedByUser := make(map[string]map[string]int)

	if len(embeds) > 0 {
		codeSubmissions, err := s.repo.GetLatestScoresByMaterialsForSection(ctx, embeds, sectionID, labID)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, sub := range codeSubmissions {
			autoScores[sub.UserID] += sub.AutoScore
			if submittedItems[sub.UserID] == nil {
				submittedItems[sub.UserID] = make(map[string]bool)
			}
			submittedItems[sub.UserID][sub.MaterialID] = true
			if sub.Status == models.PASSED {
				passedCount[sub.UserID]++
			}
			if perEmbedByUser[sub.UserID] == nil {
				perEmbedByUser[sub.UserID] = make(map[string]int)
			}
			perEmbedByUser[sub.UserID][sub.MaterialID] = sub.AutoScore
		}
	}

	if len(inputNodes) > 0 {
		inputSubs, err := s.inputSubmissionRepo.ListLatestByMaterialSectionLab(ctx, materialID, sectionID, labID)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, sub := range inputSubs {
			autoScores[sub.UserID] += sub.Score
			if submittedItems[sub.UserID] == nil {
				submittedItems[sub.UserID] = make(map[string]bool)
			}
			submittedItems[sub.UserID][sub.NodeID] = true
			// A manual submission awaiting grading (graded=false) is not "passed"
			// yet, so it keeps the document status at PARTIAL until graded.
			if sub.Passed && sub.Graded {
				passedCount[sub.UserID]++
			}
		}
	}

	statusByUser := make(map[string]models.SubmissionStatus)
	total := len(embeds) + len(inputNodes)
	for userID, submitted := range submittedItems {
		passed := passedCount[userID]
		switch {
		case passed == total:
			statusByUser[userID] = models.PASSED
		case len(submitted) > 0:
			statusByUser[userID] = models.PARTIAL
		default:
			statusByUser[userID] = models.QUEUED
		}
	}

	return autoScores, statusByUser, perEmbedByUser, nil
}

// docInputNode is an input embed node's identity and configured max score.
type docInputNode struct {
	NodeID string
	Score  int
}

func collectInputNodes(nodes []inputTiptapNode) []docInputNode {
	var out []docInputNode
	for i := range nodes {
		n := &nodes[i]
		if n.Type == "inputEmbed" {
			id, _ := n.Attrs["nodeID"].(string)
			scoreF, _ := n.Attrs["score"].(float64)
			if id != "" {
				out = append(out, docInputNode{NodeID: id, Score: int(scoreF)})
			}
		}
		out = append(out, collectInputNodes(n.Content)...)
	}
	return out
}

// docInputNodes parses a document material's content and returns its input embed
// nodes. Returns nil for a missing row, empty content, or unparseable JSON.
func (s *submissionService) docInputNodes(ctx context.Context, materialID string) ([]docInputNode, error) {
	doc, err := s.docMatRepo.GetByID(ctx, materialID)
	if err != nil {
		var csErr *cserrors.Error
		if errors.As(err, &csErr) && csErr.HttpStatus == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}
	if doc.Content == nil {
		return nil, nil
	}
	var root inputTiptapNode
	if err := json.Unmarshal([]byte(*doc.Content), &root); err != nil {
		return nil, nil
	}
	return collectInputNodes(root.Content), nil
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
		rawSubmissions, err = s.typingSubmissionRepo.GetBestByMaterial(ctx, materialID, labID, sectionID)
		if err != nil {
			return nil, err
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

	// For document materials, compute auto_score, status, and per-embed scores
	// from embedded code submissions.
	var embedAutoScores map[string]int
	var embedStatusByUser map[string]models.SubmissionStatus
	var embedPerUser map[string]map[string]int
	if mat.Type == "document" {
		embedAutoScores, embedStatusByUser, embedPerUser, _ = s.docEmbedScores(ctx, materialID, sectionID, labID)
	}

	result := make([]models.CMSSectionStudentSubmission, len(students))
	for i, student := range students {
		sub, hasSubmission := submissionMap[student.ID]

		// Build per-embed score payload for document materials so the frontend
		// can render individual embed scores in the submission detail panel.
		var docPayload any
		if mat.Type == "document" {
			embedScores := map[string]int{}
			if embedPerUser != nil {
				for matID, score := range embedPerUser[student.ID] {
					embedScores[matID] = score
				}
			}
			docPayload = map[string]any{"embed_scores": embedScores}
		}

		if hasSubmission {
			payload, err := handler.Get(ctx, sub.ID, "instructor")
			if err != nil {
				return nil, err
			}
			if mat.Type == "document" {
				payload = docPayload
			}

			autoScore := sub.AutoScore
			status := sub.Status
			if mat.Type == "document" {
				if score, ok := embedAutoScores[student.ID]; ok {
					autoScore = score
				}
				if st, ok := embedStatusByUser[student.ID]; ok {
					status = st
				}
			}

			result[i] = models.CMSSectionStudentSubmission{
				Student: student,
				StudentSubmission: &models.StudentSubmission{
					Order:       sub.Order,
					AutoScore:   autoScore,
					ManualScore: sub.ManualScore,
					IP:          sub.IPAddress,
					Status:      status,
					CreatedAt:   sub.CreatedAt,
					UpdatedAt:   sub.UpdatedAt,
					Payload:     payload,
				},
			}
		} else {
			// Student has no document submission but may have embedded code submissions.
			autoScore := 0
			status := models.NOT_SUBMITTED
			if mat.Type == "document" {
				if score, ok := embedAutoScores[student.ID]; ok {
					autoScore = score
				}
				if st, ok := embedStatusByUser[student.ID]; ok {
					status = st
				}
			}
			result[i] = models.CMSSectionStudentSubmission{
				Student: student,
				StudentSubmission: &models.StudentSubmission{
					AutoScore:   autoScore,
					ManualScore: 0,
					IP:          "",
					Status:      status,
					Payload:     docPayload,
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

	isDocument := mat.Type == "document"

	// For document materials, compute the current aggregate auto_score, status,
	// and per-embed scores from the student's embedded code submissions.
	var embedAutoScore int
	var embedStatus models.SubmissionStatus
	var embedPerEmbed map[string]int
	if isDocument {
		embedAutoScores, embedStatuses, embedPerUser, _ := s.docEmbedScores(ctx, materialID, sectionID, labID)
		if embedAutoScores != nil {
			embedAutoScore = embedAutoScores[studentID]
		}
		if embedStatuses != nil {
			embedStatus = embedStatuses[studentID]
		}
		if embedPerUser != nil {
			embedPerEmbed = embedPerUser[studentID]
		}
	}

	// For document materials with no direct submission row, synthesize one from
	// the embed state so the CMS "view all submissions" panel has something to show.
	if len(rawSubmissions) == 0 {
		if isDocument && embedStatus != "" {
			embedScores := map[string]int{}
			for matID, score := range embedPerEmbed {
				embedScores[matID] = score
			}
			status := embedStatus
			return []models.StudentSubmission{
				{
					Status:    status,
					AutoScore: embedAutoScore,
					Payload:   map[string]any{"embed_scores": embedScores},
				},
			}, 1, nil
		}
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
		autoScore := sub.AutoScore
		status := sub.Status
		payload := payloads[sub.ID]
		if isDocument {
			autoScore = embedAutoScore
			if embedStatus != "" {
				status = embedStatus
			}
			embedScores := map[string]int{}
			for matID, score := range embedPerEmbed {
				embedScores[matID] = score
			}
			payload = map[string]any{"embed_scores": embedScores}
		}
		result[i] = models.StudentSubmission{
			ID:          sub.ID,
			Status:      status,
			Order:       sub.Order,
			CreatedAt:   sub.CreatedAt,
			UpdatedAt:   sub.UpdatedAt,
			AutoScore:   autoScore,
			ManualScore: sub.ManualScore,
			IP:          sub.IPAddress,
			Payload:     payload,
		}
	}

	return result, count, nil
}

func (s *submissionService) DeleteSubmission(ctx context.Context, submissionID string) error {
	return s.repo.Delete(ctx, submissionID)
}

// regradeConcurrency caps how many submissions a single "regrade all" request
// queues in parallel, bounding the DB connection pressure from one section.
const regradeConcurrency = 10

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
	// after which Fiber recycles the underlying *fasthttp.RequestCtx and reuses it
	// for the next request. These goroutines outlive the request, so anything still
	// referencing that ctx (even via context.WithoutCancel, whose Value/Deadline
	// still delegate to the recycled parent) reads reset/foreign state — a data
	// race. Use a fresh background context, matching the detach pattern used
	// elsewhere in the codebase.
	bgCtx := context.Background()

	// Fan out under a bounded pool so a large section (hundreds of students) can't
	// open one DB transaction per submission at once and exhaust the connection
	// pool. The whole fan-out runs in a detached goroutine so the handler still
	// returns immediately; concurrency is capped inside.
	go func() {
		sem := make(chan struct{}, regradeConcurrency)
		var wg sync.WaitGroup

		for _, sub := range submissions {
			sub := sub
			sem <- struct{}{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() { <-sem }()

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

		wg.Wait()
	}()

	return nil
}
