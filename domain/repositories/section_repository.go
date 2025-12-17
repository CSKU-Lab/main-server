package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type SectionRepository interface {
	Create(ctx context.Context, ID string, section *CreateSection) error
	UpdateByID(ctx context.Context, ID string, section *UpdateSection) error
	GetByID(ctx context.Context, ID string) (*Section, error)
	GetBySemesterID(ctx context.Context, semesterID string) ([]models.Section, error)
	GetByCourseID(ctx context.Context, courseID string) ([]models.Section, error)
	GetByCourseIDAndSemesterID(ctx context.Context, courseID string, semesterID string) ([]models.Section, error)
	GetRawBySemesterID(ctx context.Context, semesterID string) ([]Section, error)
	DeleteByID(ctx context.Context, ID string) error
	DeleteByCourseID(ctx context.Context, courseID string) error
}

type CreateSection struct {
	Name       string
	CourseID   string
	SemesterID string
}

type UpdateSection struct {
	Name       string
	SemesterID string
	Banner     *string
}

type Section struct {
	ID         string
	Name       string
	Banner     *string
	SemesterID string
	CourseID   string
}
