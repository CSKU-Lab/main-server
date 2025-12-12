package services

import (
	"context"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/google/uuid"
)

type LabService interface {
	GetByID(ctx context.Context, labID string) (*models.Lab, error)
	Create(ctx context.Context, req *requests.CreateLab, userID string) (string, error)
}

type labServcie struct {
	labRepo    repositories.LabRepository
	courseRepo repositories.CourseRepository
	uowRepo    repositories.UoWRepository
}

func NewLabService(
	labRepo repositories.LabRepository,
	courseRepo repositories.CourseRepository,
	uowRepo repositories.UoWRepository,
) LabService {
	return &labServcie{
		labRepo:    labRepo,
		courseRepo: courseRepo,
		uowRepo:    uowRepo,
	}
}

func (l *labServcie) GetByID(ctx context.Context, labID string) (*models.Lab, error) {
	lab, err := l.labRepo.GetByID(ctx, labID)
	if err != nil {
		return nil, err
	}
	return lab, nil
}

func (l *labServcie) Create(ctx context.Context, req *requests.CreateLab, userID string) (string, error) {
	_, err := l.courseRepo.GetByID(ctx, req.CourseID)
	if err != nil {
		return "", err
	}

	var labID string
	err = l.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		id, err := uuid.NewV7()
		if err != nil {
			return cserrors.New(&cserrors.Option{
				Message:    "Failed to generate UUID",
				HttpStatus: http.StatusInternalServerError,
			})
		}

		err = u.Lab().Create(ctx, id.String(), req, userID)
		if err != nil {
			return err
		}
		labID = id.String()

		return nil
	})
	if err != nil {
		return "", err
	}

	return labID, nil
}
