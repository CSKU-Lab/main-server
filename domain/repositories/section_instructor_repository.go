package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type SectionInstructorRepository interface {
	Get(ctx context.Context, sectionID string) ([]models.SectionInstructor, error)
	Add(ctx context.Context, sectionID string, id string) error
	DeleteBySectionID(ctx context.Context, sectionID string) error
}
