package sqlx

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type codeSubmission struct {
	db instance
}

func NewCodeSubmission(db instance) repositories.CodeSubmission {
	return &codeSubmission{
		db: db,
	}
}

func (c *codeSubmission) Create(ctx context.Context, payload *repositories.CreateCodeSubmissionPayload) error {
	query := `INSERT INTO code_submissions (submission_id, code) VALUES ($1, $2)`

	_, err := c.db.ExecContext(ctx, query, payload.SubmissionID, payload.Code)
	if err != nil {
		return err
	}

	return nil
}
func (c *codeSubmission) Update(ctx context.Context, payload *repositories.UpdateCodeSubmissionPayload) error {
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

func (c *codeSubmission) Get(ctx context.Context, id string) (*models.CodeSubmission, error) {
	return nil, nil
}
