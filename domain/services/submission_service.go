package services

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	// err := s.checkPermission(ctx, submissionID)
	// if err != nil {
	// 	return nil, err
	// }

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
		ID:      submission.ID,
		UserID:  submission.UserID,
		Status:  submission.Status,
		Payload: payload,
	}, nil
}

func (s *submissionService) Listen(ctx context.Context, submissionID string) (<-chan *models.Submission, error) {
	// err := s.checkPermission(ctx, submissionID)
	// if err != nil {
	// 	return nil, err
	// }
	submission, err := s.Get(ctx, submissionID)
	if err != nil {
		return nil, err
	}

	subChan := make(chan *models.Submission)
	if submission.Status == models.PASSED || submission.Status == models.FAILED {
		close(subChan)
		return subChan, nil
	}

	channel := fmt.Sprintf("submissions:update:%s", submissionID)
	go s.ps.Subscribe(ctx, channel, func(payload []byte) error {
		submission, err := s.Get(ctx, submissionID)
		if err != nil {
			close(subChan)
		}
		subChan <- submission
		log.Println(submission)

		status := string(payload)
		if status == "failed" || status == "passed" {
			close(subChan)
			return pubsub.Exit
		}

		return nil
	})

	return subChan, nil
}

func (s *submissionService) checkPermission(ctx context.Context, id string) error {
	user, ok := ctx.Value("user").(*models.User)
	if !ok {
		return errors.New("cannot get user")
	}

	submission, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if submission.UserID != user.ID {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Code:       cserrors.Forbidden,
			Message:    "You do not have access to this submission",
		})
	}
	return nil
}
