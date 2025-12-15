package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type SectionStudentRepository interface {
	Add(ctx context.Context, sectionID string, studentID string) error
	GetBySectionID(ctx context.Context, sectionID string) ([]models.Student, error)
	RemoveBySectionIDAndStudentID(ctx context.Context, sectionID string, studentID string) error
}
