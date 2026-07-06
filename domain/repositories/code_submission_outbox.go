package repositories

import (
	"context"
	"time"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type CodeSubmissionOutboxRepository interface {
	Create(ctx context.Context, id string, submisionID string, payload *models.GradeExecution) error
	Update(ctx context.Context, id string, isSent bool) error
	Get(ctx context.Context, id string) (*CodeSubmissionOutboxPayload, error)
	Delete(ctx context.Context, id string) error
	GetUnsent(ctx context.Context, limit int, olderThan time.Duration) ([]*CodeSubmissionOutboxPayload, error)
	// ClaimForProcessing atomically claims a record for one worker instance:
	// it succeeds for exactly one caller while the record is unsent, not yet
	// dead-lettered (retry_count < 3), and not claimed within staleAfter. It
	// bumps retry_count and last_attempt_at so a claim whose consumer dies
	// becomes reclaimable after staleAfter (the grade result is otherwise
	// orphaned). Only the claimer publishes the grade and consumes the result.
	ClaimForProcessing(ctx context.Context, id string, staleAfter time.Duration) (bool, error)
	// MarkSent marks a record done — call ONLY after the terminal grade result
	// has been consumed and the submission status persisted.
	MarkSent(ctx context.Context, id string) error
	IncrementRetry(ctx context.Context, id string) error
}

type CodeSubmissionOutboxPayload struct {
	ID           string
	SubmissionID string
	IsSent       bool
	Payload      string
	CreatedAt    time.Time
}
