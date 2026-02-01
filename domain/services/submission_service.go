package services

import (
	"context"
	"errors"

	contextkeys "github.com/CSKU-Lab/main-server/context_keys"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/google/uuid"
)

type SubmissionService interface {
	Create(ctx context.Context, req *requests.Submission, rawPayload []byte) (string, error)
	Update(ctx context.Context, submissionID string, payload *UpdateSubmissionPayload, rawPayload []byte) error
	Get(ctx context.Context, ID string) (*models.Submission, error)
}

type UpdateSubmissionPayload struct {
	Type    string
	Status  models.SubmissionStatus
	Payload any `json:"payload"`
}

type submissionService struct {
	repo     repositories.Submission
	uowRepo  repositories.UoWRepository
	registry registries.SubmissionRegistry
}

func NewSubmissionService(repo repositories.Submission, uowRepo repositories.UoWRepository, registry registries.SubmissionRegistry) SubmissionService {
	return &submissionService{
		repo:     repo,
		uowRepo:  uowRepo,
		registry: registry,
	}
}

func (s *submissionService) Create(ctx context.Context, req *requests.Submission, rawPayload []byte) (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}

	user, ok := ctx.Value(contextkeys.UserKey).(contextkeys.User)
	if !ok {
		return "", errors.New("cannot extract user from context")
	}

	payload := &repositories.SubmissionPayload{
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

		handler, err := s.registry.GetHandler(req.Type)
		if err != nil {
			return err
		}

		return handler.Create(ctx, u, id.String(), rawPayload)
	})
	if err != nil {
		return "", err
	}

	return id.String(), nil
}

func (s *submissionService) Update(ctx context.Context, submissionID string, payload *UpdateSubmissionPayload, rawPayload []byte) error {
	return s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		err := u.Submission().Update(ctx, submissionID, payload.Status)
		if err != nil {
			return err
		}

		handler, err := s.registry.GetHandler(payload.Type)
		if err != nil {
			return err
		}

		return handler.Update(ctx, u, submissionID, rawPayload)
	})
}

func (s *submissionService) Get(ctx context.Context, id string) (*models.Submission, error) {
	submission, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	handler, err := s.registry.GetHandler(submission.Type)
	if err != nil {
		return nil, err
	}

	payload, err := handler.Get(ctx, submission.ID)
	if err != nil {
		return nil, err
	}

	return &models.Submission{
		ID:      submission.ID,
		Type:    submission.Type,
		Status:  submission.Status,
		Payload: payload,
	}, nil
}
