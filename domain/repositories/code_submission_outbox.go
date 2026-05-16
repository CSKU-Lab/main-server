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
	TryMarkSent(ctx context.Context, id string) (bool, error)
	IncrementRetry(ctx context.Context, id string) error
}

type CodeSubmissionOutboxPayload struct {
	ID           string
	SubmissionID string
	IsSent       bool
	Payload      string
	CreatedAt    time.Time
}
