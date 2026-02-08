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
	ID         string                  `db:"id"`
	Status     models.SubmissionStatus `db:"status"`
	UserID     string                  `db:"user_id"`
	LabID      string                  `db:"lab_id"`
	MaterialID string                  `db:"material_id"`
	SectionID  *string                 `db:"section_id"`
	CourseID   *string                 `db:"course_id"`
	CreatedAt  time.Time               `db:"created_at"`
	UpdatedAt  time.Time               `db:"updated_at"`
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
		Status:     submission.Status,
	}

	return model, nil
}

type submissionOverview struct {
	ID             string                      `db:"id"`
	Status         models.SubmissionStatus     `db:"status"`
	CreatedAt      time.Time                   `db:"created_at"`
	TestCaseGroups models.TestCaseGroupResults `db:"test_case_groups"`
}

func (s *submissionRepository) GetPagination(ctx context.Context, userID string, materialID string, page int, pageSize int, sortOrder string) ([]models.SubmissionOverview, error) {
	query := `SELECT s.id, s.status, s.created_at,
			  cs.test_case_groups
			  FROM submissions s
			  LEFT JOIN code_submissions cs ON s.id = cs.submission_id
			  WHERE s.user_id = $1 AND s.material_id = $2
			  ORDER BY s.created_at %s
			  OFFSET $3 LIMIT $4`

	query = fmt.Sprintf(query, sortOrder)
	offset := (page - 1) * pageSize

	rows := []submissionOverview{}
	err := s.db.SelectContext(ctx, &rows, query, userID, materialID, offset, pageSize)
	if err != nil {
		return nil, err
	}

	result := make([]models.SubmissionOverview, len(rows))
	for i, row := range rows {
		overview := models.SubmissionOverview{
			ID:        row.ID,
			Status:    row.Status,
			CreatedAt: row.CreatedAt,
		}

		// Calculate test case counts from JSONB test_case_groups
		if len(row.TestCaseGroups) > 0 {
			totalTestCases := 0
			passedTestCases := 0

			for _, group := range row.TestCaseGroups {
				totalTestCases += len(group.Results)
				for _, tc := range group.Results {
					if tc.Status == models.CODE_EXECUTION_RUN_PASSED {
						passedTestCases++
					}
				}
			}

			overview.Payload = models.CodeSubmissionOverviewPayload{
				TotalTestCases:  totalTestCases,
				PassedTestCases: passedTestCases,
			}
		}

		result[i] = overview
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
