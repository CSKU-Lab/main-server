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
}

type SubmissionServiceArgs struct {
	SubmissionRepository     repositories.SubmissionRepository
	MaterialRepository       repositories.MaterialRepository
	UowRepository            repositories.UoWRepository
	SubmissionRegistry       registries.SubmissionRegistry
	SectionStudentRepository repositories.SectionStudentRepository
}

func NewSubmissionService(args *SubmissionServiceArgs) SubmissionService {
	return &submissionService{
		repo:               args.SubmissionRepository,
		materialRepo:       args.MaterialRepository,
		uowRepo:            args.UowRepository,
		registry:           args.SubmissionRegistry,
		sectionStudentRepo: args.SectionStudentRepository,
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
			return "", err
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

		return handler.Create(ctx, u, id.String(), rawPayload)
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

func (s *submissionService) Get(ctx context.Context, id string) (*models.Submission, error) {
	submission, err := s.repo.Get(ctx, id)
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
		ID:      submission.ID,
		Payload: payload,
	}, nil
}
