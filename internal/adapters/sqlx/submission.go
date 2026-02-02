package sqlx

import (
	"context"
	"time"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type submissionRepository struct {
	db instance
}

func NewSubmissionRepository(db instance) repositories.SubmissionRepository {
	return &submissionRepository{
		db: db,
	}
}

func (s *submissionRepository) Create(ctx context.Context, payload *repositories.Submission) error {
	query := `INSERT INTO submissions
	(id, user_id, material_id, section_id, course_id, status, created_at, updated_at)
	VALUES ($1,$2,$3,$4,$5,'queued',NOW(),NOW())`

	_, err := s.db.ExecContext(ctx, query, payload.ID, payload.UserID, payload.MaterialID, payload.SectionID, payload.CourseID)
	if err != nil {
		return err
	}
	return nil
}

func (s *submissionRepository) Update(ctx context.Context, id string, status models.SubmissionStatus) error {
	query := `UPDATE submissions
	SET status = $2, updated_at = NOW() WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, id, string(status))
	if err != nil {
		return err
	}
	return nil
}

type submission struct {
	ID         string    `db:"id"`
	Status     string    `db:"status"`
	UserID     string    `db:"user_id"`
	LabID      string    `db:"lab_id"`
	MaterialID string    `db:"material_id"`
	SectionID  *string   `db:"section_id"`
	CourseID   *string   `db:"course_id"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func (s *submissionRepository) Get(ctx context.Context, id string) (*repositories.Submission, error) {
	query := `SELECT * FROM submissions WHERE id = $1`

	submission := submission{}
	err := s.db.GetContext(ctx, &submission, query, id)
	if err != nil {
		return nil, err
	}

	model := &repositories.Submission{
		ID:         submission.ID,
		UserID:     submission.UserID,
		LabID:      submission.LabID,
		SectionID:  submission.SectionID,
		CourseID:   submission.CourseID,
		MaterialID: submission.MaterialID,
	}

	return model, nil
}
