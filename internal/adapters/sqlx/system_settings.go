package sqlx

import (
	"context"
	"database/sql"
	"errors"

	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type systemSettingsRepository struct {
	db instance
}

func NewSystemSettingsRepository(db instance) repositories.SystemSettingsRepository {
	return &systemSettingsRepository{db: db}
}

func (r *systemSettingsRepository) Get(ctx context.Context, key string) (*string, error) {
	var value string
	err := r.db.GetContext(ctx, &value, `SELECT value FROM system_settings WHERE key = $1`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (r *systemSettingsRepository) Set(ctx context.Context, key, value string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO system_settings (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`, key, value)
	return err
}
