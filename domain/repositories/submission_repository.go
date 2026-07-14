package repositories

import (
	"context"
	"time"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type UpdateSubmissionRequest struct {
	ID          string
	Status      *models.SubmissionStatus
	AutoScore   *int
	ManualScore *int
}

type SubmissionRepository interface {
	Create(ctx context.Context, req *Submission) error
	Update(ctx context.Context, req *UpdateSubmissionRequest) error
	Get(ctx context.Context, id string) (*Submission, error)
	GetPagination(ctx context.Context, userID string, materialID string, labID string, sectionID string, page int, pageSize int, sortOrder string) ([]Submission, error)
	GetLatestOfStudentIDInSectionID(ctx context.Context, sectionID, labID, materialID, studentID string) (*models.RawSubmission, error)
	GetLatestByMaterialSectionAndLabID(ctx context.Context, materialID string, sectionID string, labID string) ([]models.RawSubmission, error)
	GetPaginationByMaterialSectionLabAndStudentID(ctx context.Context, materialID string, sectionID string, labID string, studentID string, page int, pageSize int, sortBy, sortOrder string) ([]models.RawSubmission, error)

	Count(ctx context.Context, userID string, materialID string, labID string, sectionID string) (int, error)
	CountByMaterialSectionLabAndStudentID(ctx context.Context, materialID string, sectionID string, labID string, studentID string) (int, error)

	Delete(ctx context.Context, id string) error
	// GetLatestScoresByMaterialsForSection returns the latest submission (by order)
	// per (user_id, material_id) for each given material in the section/lab.
	// Used to aggregate embedded code problem scores into document submissions.
	GetLatestScoresByMaterialsForSection(ctx context.Context, materialIDs []string, sectionID, labID string) ([]models.RawSubmission, error)
	// GetLatestScoresBySection returns the latest submission (by order) per
	// (user_id, lab_id, material_id) across the whole section in one query.
	// Used by the gradebook to avoid a per-cell lookup.
	GetLatestScoresBySection(ctx context.Context, sectionID string) ([]models.RawSubmission, error)
}

type Submission struct {
	ID         string
	UserID     string
	LabID      string
	SectionID  *string
	CourseID   *string
	MaterialID string
	Status     models.SubmissionStatus
	Order      int
	AutoScore  int
	CreatedAt  time.Time
}
