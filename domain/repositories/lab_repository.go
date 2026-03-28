package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

type LabRepository interface {
	GetByID(ctx context.Context, labID string) (*models.Lab, error)
	Create(ctx context.Context, id string, req *requests.CreateLab, userID string) error
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter) ([]models.Lab, error)
	Count(ctx context.Context, search string, filters []sanitize.Filter) (int, error)
	UpdateByID(ctx context.Context, labID string, req *requests.BaseUpdateLab) error
	DeleteByID(ctx context.Context, labID string) error
	// ExistsByNameInCourse checks if a lab with the given display_name exists in the specified course
	ExistsByNameInCourse(ctx context.Context, displayName string, courseID string) (bool, error)
	// ExistsByNameInCourseExcludingID checks if a lab with the given display_name exists in the specified course, excluding the given lab ID (for updates)
	ExistsByNameInCourseExcludingID(ctx context.Context, displayName string, courseID string, excludeLabID string) (bool, error)
}
