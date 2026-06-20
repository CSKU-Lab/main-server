package repositories

import "context"

type SystemSettingsRepository interface {
	Get(ctx context.Context, key string) (*string, error)
	Set(ctx context.Context, key, value string) error
}
