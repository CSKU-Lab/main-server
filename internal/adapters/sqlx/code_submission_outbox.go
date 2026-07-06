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
	RetryCount  int       `db:"retry_count"`
	LastAttempt *time.Time `db:"last_attempt_at"`
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

func (c *codeSubmissionOutboxRepository) GetUnsent(ctx context.Context, limit int, olderThan time.Duration) ([]*repositories.CodeSubmissionOutboxPayload, error) {
	// Surface records not attempted within olderThan (last_attempt_at, NULL for
	// never-attempted) so reconciliation retries claims whose consumer died, not
	// just brand-new records.
	query := `SELECT * FROM code_submissions_outbox WHERE is_sent = false AND retry_count < 3 AND (last_attempt_at IS NULL OR last_attempt_at < NOW() - ($1 || ' seconds')::interval) ORDER BY created_at ASC LIMIT $2`

	var records []codeSubmissionOutboxRecord
	err := c.db.SelectContext(ctx, &records, query, int(olderThan.Seconds()), limit)
	if err != nil {
		return nil, err
	}

	result := make([]*repositories.CodeSubmissionOutboxPayload, 0, len(records))
	for _, r := range records {
		result = append(result, &repositories.CodeSubmissionOutboxPayload{
			ID:           r.ID,
			SubmissionID: r.SubmissinID,
			IsSent:       r.IsSent,
			Payload:      r.Payload,
			CreatedAt:    r.CreatedAt,
		})
	}

	return result, nil
}

// ClaimForProcessing atomically claims an unsent, not-dead-lettered record that
// has not been attempted within staleAfter. Exactly one instance wins; it bumps
// retry_count + last_attempt_at so a claim whose consumer dies is reclaimable
// after staleAfter instead of orphaning the grade result.
func (c *codeSubmissionOutboxRepository) ClaimForProcessing(ctx context.Context, id string, staleAfter time.Duration) (bool, error) {
	query := `UPDATE code_submissions_outbox
		SET retry_count = retry_count + 1, last_attempt_at = NOW()
		WHERE id = $1 AND is_sent = false AND retry_count < 3
		  AND (last_attempt_at IS NULL OR last_attempt_at < NOW() - ($2 || ' seconds')::interval)`

	res, err := c.db.ExecContext(ctx, query, id, int(staleAfter.Seconds()))
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

// MarkSent marks a record done. Call ONLY after the terminal grade result has
// been consumed and the submission status persisted.
func (c *codeSubmissionOutboxRepository) MarkSent(ctx context.Context, id string) error {
	query := `UPDATE code_submissions_outbox SET is_sent = true, last_attempt_at = NOW() WHERE id = $1`
	_, err := c.db.ExecContext(ctx, query, id)
	return err
}

func (c *codeSubmissionOutboxRepository) IncrementRetry(ctx context.Context, id string) error {
	query := `UPDATE code_submissions_outbox SET retry_count = retry_count + 1, last_attempt_at = NOW() WHERE id = $1`
	_, err := c.db.ExecContext(ctx, query, id)
	return err
}
