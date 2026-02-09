package sqlx

import (
	"context"
	"fmt"
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
	(id, user_id, material_id, lab_id, section_id, course_id, status, submission_order, created_at, updated_at)
	SELECT $1, $2, $3, $4, $5, $6, 'queued', COALESCE(MAX(submission_order), 0) + 1, NOW(), NOW()
	FROM submissions
	WHERE user_id = $2 AND material_id = $3`

	_, err := s.db.ExecContext(ctx, query, payload.ID, payload.UserID, payload.MaterialID, payload.LabID, payload.SectionID, payload.CourseID)
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
	ID         string                  `db:"id"`
	Status     models.SubmissionStatus `db:"status"`
	UserID     string                  `db:"user_id"`
	LabID      string                  `db:"lab_id"`
	MaterialID string                  `db:"material_id"`
	SectionID  *string                 `db:"section_id"`
	CourseID   *string                 `db:"course_id"`
	Order      int                     `db:"submission_order"`
	CreatedAt  time.Time               `db:"created_at"`
	UpdatedAt  time.Time               `db:"updated_at"`
}

func (s *submissionRepository) Get(ctx context.Context, id string) (*repositories.Submission, error) {
	query := `SELECT id, user_id, lab_id, section_id, course_id, material_id, status, submission_order, created_at
              FROM submissions
              WHERE id = $1`

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
		Status:     submission.Status,
		Order:      submission.Order,
		CreatedAt:  submission.CreatedAt,
	}

	return model, nil
}

func (s *submissionRepository) GetPagination(ctx context.Context, userID string, materialID string, page int, pageSize int, sortOrder string) ([]repositories.Submission, error) {
	query := `SELECT id, user_id, lab_id, section_id, course_id, material_id, status, submission_order, created_at
			  FROM submissions
			  WHERE user_id = $1 AND material_id = $2
			  ORDER BY created_at %s
			  OFFSET $3 LIMIT $4`

	query = fmt.Sprintf(query, sortOrder)
	offset := (page - 1) * pageSize

	rows := []submission{}
	err := s.db.SelectContext(ctx, &rows, query, userID, materialID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	result := make([]repositories.Submission, len(rows))
	for i, row := range rows {
		result[i] = repositories.Submission{
			ID:         row.ID,
			UserID:     row.UserID,
			LabID:      row.LabID,
			SectionID:  row.SectionID,
			CourseID:   row.CourseID,
			MaterialID: row.MaterialID,
			Status:     row.Status,
			Order:      row.Order,
			CreatedAt:  row.CreatedAt,
		}
	}

	return result, nil
}

func (s *submissionRepository) Count(ctx context.Context, userID string, materialID string) (int, error) {
	query := `SELECT COUNT(*) FROM submissions WHERE user_id = $1 AND material_id = $2`

	var count int
	err := s.db.GetContext(ctx, &count, query, userID, materialID)
	if err != nil {
		return 0, err
	}

	return count, nil
}
