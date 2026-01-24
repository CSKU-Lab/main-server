package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

type SectionStudentRepository interface {
	Add(ctx context.Context, sectionID string, studentID string) error
	GetBySectionID(ctx context.Context, sectionID string) ([]models.Student, error)
	GetSectionsPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filters []sanitize.Filter) ([]RawSection, error)
	GetBySectionAndStudentID(ctx context.Context, sectionID string, studentID string) (*models.SectionStudent, error)
	GetByStudentID(ctx context.Context, studentID string) ([]models.Section, error)
	Count(ctx context.Context, filters []sanitize.Filter) (int, error)
	RemoveBySectionIDAndStudentID(ctx context.Context, sectionID string, studentID string) error
}
