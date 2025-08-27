package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type CourseCreatorRepository interface {
	GetCreators(ctx context.Context, courseID string) ([]models.CourseCreator, error)
	SetCreators(ctx context.Context, courseID string, creators []string) error
}
