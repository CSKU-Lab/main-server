package sqlx

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type authLogRepository struct {
	db instance
}

func NewAuthLogRepository(db instance) repositories.AuthLogRepository {
	return &authLogRepository{db: db}
}

func (r *authLogRepository) Create(ctx context.Context, id string, userID string, action string) error {
	query := `INSERT INTO auth_logs (id, user_id, action, created_at) VALUES ($1, $2, $3, now())`

	_, err := r.db.ExecContext(ctx, query, id, userID, action)
	if err != nil {
		return err
	}

	return nil
}
