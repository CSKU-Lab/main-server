package sqlx

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type userActivityRepository struct {
	db instance
}

func NewUserActivityRepository(db instance) repositories.UserActivityRepository {
	return &userActivityRepository{db: db}
}

func (r *userActivityRepository) Touch(ctx context.Context, userID string) error {
	query := `
		INSERT INTO user_activity (user_id, last_seen)
		VALUES ($1, now())
		ON CONFLICT (user_id) DO UPDATE SET last_seen = now()
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}
