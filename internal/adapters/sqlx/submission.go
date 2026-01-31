package sqlx

import (
	"context"
	"fmt"
	"time"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
)

type submissionRepository struct {
	db *sqlx.DB
}

func NewSubmissionRepository(db *sqlx.DB) repositories.Submission {
	return &submissionRepository{
		db: db,
	}
}

type submission struct {
	ID         string    `db:"id"`
	Status     string    `db:"status"`
	UserID     string    `db:"user_id"`
	MaterialID string    `db:"material_id"`
	SectionID  *string   `db:"section_id"`
	CourseID   *string   `db:"course_id"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func (s *submissionRepository) Create(ctx context.Context, payload *repositories.SubmissionPayload) error {
	query := `INSERT INTO submissions
	(id, user_id, material_id, section_id, course_id, created_at)
	VALUES ($1,$2,$3,$4,$5,NOW())`

	_, err := s.db.ExecContext(ctx, query, payload.ID, payload.UserID, payload.MaterialID, payload.SectionID, payload.CourseID)
	if err != nil {
		return err
	}
	return nil
}

func (s *submissionRepository) Update(ctx context.Context, payload *repositories.SubmissionPayload) error {
	fields := &submission{
		ID:         payload.ID,
		UserID:     payload.UserID,
		MaterialID: payload.MaterialID,
		SectionID:  payload.SectionID,
		CourseID:   payload.CourseID,
	}

	updateFields := getUpdateFields(fields)
	if len(updateFields) == 0 {
		return nil
	}

	query := fmt.Sprintf(`UPDATE submissions
	SET %s, updated_at = NOW()`, updateFields)

	_, err := s.db.ExecContext(ctx, query, fields)
	if err != nil {
		return err
	}
	return nil
}

func (s *submissionRepository) Get(ctx context.Context, ID string) (*models.Submission, error) {
	query := `SELECT * FROM submissions WHERE id = $1`

	submission := &submission{}
	err := s.db.GetContext(ctx, &submission, query)
	if err != nil {
		return nil, err
	}

	model := &models.Submission{
		ID:     submission.ID,
		Status: models.SubmissionStatus(submission.Status),
	}

	return model, nil
}
