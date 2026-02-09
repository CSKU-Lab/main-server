package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	contextkeys "github.com/CSKU-Lab/main-server/context_keys"
	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/adapters/pubsub"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/google/uuid"
)

type SubmissionService interface {
	Create(ctx context.Context, req *requests.Submission, rawPayload []byte) (string, error)
	Update(ctx context.Context, submissionID string, payload *UpdateSubmissionPayload, rawPayload []byte) error
	Get(ctx context.Context, submissionID string) (*models.Submission, error)
	Listen(ctx context.Context, submissionID string) (<-chan *models.Submission, error)
	GetUserSubmissionsByMaterial(ctx context.Context, userID string, materialID string, page int, pageSize int, sortOrder string) ([]models.Submission, int, error)
}

type UpdateSubmissionPayload struct {
	Status     models.SubmissionStatus
	MaterialID string
	Payload    any `json:"payload"`
}

type submissionService struct {
	repo               repositories.SubmissionRepository
	materialRepo       repositories.MaterialRepository
	sectionStudentRepo repositories.SectionStudentRepository
	uowRepo            repositories.UoWRepository
	registry           registries.SubmissionRegistry
	ps                 pubsub.PubSub
}

type SubmissionServiceArgs struct {
	SubmissionRepository     repositories.SubmissionRepository
	MaterialRepository       repositories.MaterialRepository
	UowRepository            repositories.UoWRepository
	SubmissionRegistry       registries.SubmissionRegistry
	SectionStudentRepository repositories.SectionStudentRepository
	PubSub                   pubsub.PubSub
}

func NewSubmissionService(args *SubmissionServiceArgs) SubmissionService {
	return &submissionService{
		repo:               args.SubmissionRepository,
		materialRepo:       args.MaterialRepository,
		uowRepo:            args.UowRepository,
		registry:           args.SubmissionRegistry,
		sectionStudentRepo: args.SectionStudentRepository,
		ps:                 args.PubSub,
	}
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
		err := u.Submission().Create(ctx, payload)
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
		err := u.Submission().Update(ctx, submissionID, payload.Status)
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

	payload, err := handler.Get(ctx, submission.ID)
	if err != nil {
		return nil, err
	}

	return &models.Submission{
		ID:        submission.ID,
		Status:    submission.Status,
		Order:     submission.Order,
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

func (s *submissionService) GetUserSubmissionsByMaterial(ctx context.Context, userID string, materialID string, page int, pageSize int, sortOrder string) ([]models.Submission, int, error) {
	mat, err := s.materialRepo.GetByID(ctx, materialID)
	if err != nil {
		return nil, 0, err
	}

	submissions, err := s.repo.GetPagination(ctx, userID, materialID, page, pageSize, sortOrder)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.repo.Count(ctx, userID, materialID)
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
		result[i] = models.Submission{
			ID:        sub.ID,
			Status:    sub.Status,
			Order:     sub.Order,
			CreatedAt: sub.CreatedAt,
			Payload:   payloads[sub.ID],
		}
	}

	return result, count, nil
}
