package sqlx

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type codeSubmissionRepository struct {
	db instance
}

func NewCodeSubmission(db instance) repositories.CodeSubmissionRepository {
	return &codeSubmissionRepository{
		db: db,
	}
}

func (c *codeSubmissionRepository) Create(ctx context.Context, payload *repositories.CreateCodeSubmissionPayload) error {
	query := `INSERT INTO code_submissions (submission_id, files) VALUES ($1, $2)`

	_, err := c.db.ExecContext(ctx, query, payload.SubmissionID, payload.Files)
	if err != nil {
		return err
	}

	return nil
}
func (c *codeSubmissionRepository) Update(ctx context.Context, payload *repositories.UpdateCodeSubmissionPayload) error {
	query := `UPDATE code_submissions SET
				status = $2,
				avg_wall_time = $3,
				avg_memory = $4,
				test_case_groups = $5
			  WHERE submission_id = $1
				`

	_, err := c.db.ExecContext(ctx, query, payload.SubmissionID, payload.Status, payload.AvgWallTime, payload.AvgMemory, payload.TestCaseGroups)
	if err != nil {
		return err
	}

	return nil
}

type codeSubmission struct {
	Files          models.SubmissionFiles      `db:"files"`
	Status         *string                     `db:"status"`
	AvgWallTime    *float32                    `db:"avg_wall_time"`
	AvgMemory      *int32                      `db:"avg_memory"`
	TestCaseGroups models.TestCaseGroupResults `db:"test_case_groups"`
}

func (c *codeSubmissionRepository) Get(ctx context.Context, submissionId string) (*models.CodeSubmission, error) {
	query := `SELECT files,status,avg_wall_time,avg_memory,test_case_groups FROM code_submissions WHERE submission_id = $1`

	submission := codeSubmission{}
	err := c.db.GetContext(ctx, &submission, query, submissionId)
	if err != nil {
		return nil, err
	}

	return &models.CodeSubmission{
		Files:          submission.Files,
		Status:         submission.Status,
		AvgWallTime:    submission.AvgWallTime,
		AvgMemory:      submission.AvgMemory,
		TestCaseGroups: submission.TestCaseGroups,
	}, nil
}
