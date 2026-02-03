package sqlx

import (
	"context"
	"encoding/json"
	"time"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type codeSubmissionOutboxRepository struct {
	db instance
}

func NewCodeSubmissionOutboxRepository(db instance) repositories.CodeSubmissionOutboxRepository {
	return &codeSubmissionOutboxRepository{
		db: db,
	}
}

func (c *codeSubmissionOutboxRepository) Create(ctx context.Context, id string, submissionID string, payload *models.GradeExecution) error {
	query := `INSERT INTO code_submissions_outbox (id, submission_id, is_sent, payload, created_at) VALUES ($1,$2,$3,$4,NOW())`

	mPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(ctx, query, id, submissionID, false, mPayload)

	return err
}

func (c *codeSubmissionOutboxRepository) Update(ctx context.Context, id string, isSent bool) error {
	query := `UPDATE code_submissions_outbox SET is_sent = $1 WHERE id = $2`
	_, err := c.db.ExecContext(ctx, query, isSent, id)
	return err
}

type codeSubmissionOutboxRecord struct {
	ID          string    `db:"id"`
	SubmissinID string    `db:"submission_id"`
	IsSent      bool      `db:"is_sent"`
	Payload     string    `db:"payload"`
	CreatedAt   time.Time `db:"created_at"`
}

func (c *codeSubmissionOutboxRepository) Get(ctx context.Context, id string) (*repositories.CodeSubmissionOutboxPayload, error) {
	query := `SELECT * FROM  code_submissions_outbox WHERE id = $1`

	record := codeSubmissionOutboxRecord{}

	err := c.db.GetContext(ctx, &record, query, id)
	if err != nil {
		return nil, err
	}

	return &repositories.CodeSubmissionOutboxPayload{
		ID:           record.ID,
		SubmissionID: record.SubmissinID,
		IsSent:       record.IsSent,
		Payload:      record.Payload,
		CreatedAt:    record.CreatedAt,
	}, nil
}

func (c *codeSubmissionOutboxRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM code_submissions_outbox WHERE id = $1`
	_, err := c.db.ExecContext(ctx, query, id)
	return err
}
