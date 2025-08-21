package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type SectionRepository interface {
	Create(ctx context.Context, section *models.Section, courseID, semesterID string) error
	UpdateByID(ctx context.Context, section *models.Section) error
	GetByID(ctx context.Context, ID string) (*models.Section, error)
	GetBySemesterID(ctx context.Context, semesterID string) ([]models.Section, error)
	DeleteByID(ctx context.Context, ID string) error
}
