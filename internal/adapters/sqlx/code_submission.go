package sqlx

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
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
	SubmissionID   string                      `db:"submission_id"`
	Files          models.SubmissionFiles      `db:"files"`
	Status         *string                     `db:"status"`
	AvgWallTime    *float32                    `db:"avg_wall_time"`
	AvgMemory      *int32                      `db:"avg_memory"`
	TestCaseGroups models.TestCaseGroupResults `db:"test_case_groups"`
}

func (c *codeSubmissionRepository) Get(ctx context.Context, submissionId string) (*models.CodeSubmission, error) {
	query := `SELECT submission_id, files, status, avg_wall_time, avg_memory, test_case_groups FROM code_submissions WHERE submission_id = $1`

	submission := codeSubmission{}
	err := c.db.GetContext(ctx, &submission, query, submissionId)
	if err != nil {
		return nil, err
	}

	return &models.CodeSubmission{
		SubmissionID:   submission.SubmissionID,
		Files:          submission.Files,
		Status:         submission.Status,
		AvgWallTime:    submission.AvgWallTime,
		AvgMemory:      submission.AvgMemory,
		TestCaseGroups: submission.TestCaseGroups,
	}, nil
}

func (c *codeSubmissionRepository) GetByIDs(ctx context.Context, submissionIDs []string) ([]*models.CodeSubmission, error) {
	if len(submissionIDs) == 0 {
		return []*models.CodeSubmission{}, nil
	}

	query := `SELECT submission_id, files, status, avg_wall_time, avg_memory, test_case_groups FROM code_submissions WHERE submission_id IN (?)`

	query, args, err := sqlx.In(query, submissionIDs)
	if err != nil {
		return nil, err
	}

	query = c.db.Rebind(query)

	var submissions []codeSubmission
	err = c.db.SelectContext(ctx, &submissions, query, args...)
	if err != nil {
		return nil, err
	}

	result := make([]*models.CodeSubmission, len(submissions))
	for i, submission := range submissions {
		result[i] = &models.CodeSubmission{
			SubmissionID:   submission.SubmissionID,
			Files:          submission.Files,
			Status:         submission.Status,
			AvgWallTime:    submission.AvgWallTime,
			AvgMemory:      submission.AvgMemory,
			TestCaseGroups: submission.TestCaseGroups,
		}
	}

	return result, nil
}
