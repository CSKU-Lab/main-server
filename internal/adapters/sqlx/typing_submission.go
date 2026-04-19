package sqlx

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
)

type typingSubmissionRepo struct {
	db instance
}

func NewTypingSubmissionRepository(db instance) repositories.TypingSubmissionRepository {
	return &typingSubmissionRepo{db: db}
}

type typingSubmissionRecord struct {
	SubmissionID string  `db:"submission_id"`
	RawWPM       float64 `db:"raw_wpm"`
	AdjustedWPM  float64 `db:"adjusted_wpm"`
	ErrorRate    float64 `db:"error_rate"`
	Duration     float64 `db:"duration"`
}

func (r *typingSubmissionRepo) Create(ctx context.Context, payload *repositories.CreateTypingSubmissionPayload) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO typing_submissions (submission_id, raw_wpm, adjusted_wpm, error_rate, duration) VALUES ($1, $2, $3, $4, $5)`,
		payload.SubmissionID, payload.RawWPM, payload.AdjustedWPM, payload.ErrorRate, payload.Duration,
	)
	return err
}

func (r *typingSubmissionRepo) Get(ctx context.Context, submissionID string) (*models.TypingSubmission, error) {
	var rec typingSubmissionRecord
	err := r.db.GetContext(ctx, &rec, `SELECT raw_wpm, adjusted_wpm, error_rate, duration FROM typing_submissions WHERE submission_id = $1`, submissionID)
	if err != nil {
		return nil, err
	}
	return &models.TypingSubmission{
		RawWPM:      rec.RawWPM,
		AdjustedWPM: rec.AdjustedWPM,
		ErrorRate:   rec.ErrorRate,
		Duration:    rec.Duration,
	}, nil
}

func (r *typingSubmissionRepo) GetByIDs(ctx context.Context, submissionIDs []string) (map[string]*models.TypingSubmission, error) {
	if len(submissionIDs) == 0 {
		return map[string]*models.TypingSubmission{}, nil
	}

	query, args, err := sqlx.In(`SELECT submission_id, raw_wpm, adjusted_wpm, error_rate, duration FROM typing_submissions WHERE submission_id IN (?)`, submissionIDs)
	if err != nil {
		return nil, err
	}
	query = sqlx.Rebind(sqlx.DOLLAR, query)

	var recs []typingSubmissionRecord
	if err := r.db.SelectContext(ctx, &recs, query, args...); err != nil {
		return nil, err
	}

	result := make(map[string]*models.TypingSubmission, len(recs))
	for _, rec := range recs {
		result[rec.SubmissionID] = &models.TypingSubmission{
			RawWPM:      rec.RawWPM,
			AdjustedWPM: rec.AdjustedWPM,
			ErrorRate:   rec.ErrorRate,
			Duration:    rec.Duration,
		}
	}
	return result, nil
}
