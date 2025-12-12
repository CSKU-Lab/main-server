package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type LabRepository interface {
	GetByID(ctx context.Context, labID string) (*models.Lab, error)
	Create(ctx context.Context, id string, req *requests.CreateLab, userID string) error
}
