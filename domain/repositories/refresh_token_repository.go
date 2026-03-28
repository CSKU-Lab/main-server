package repositories

import "context"

type RefreshTokenRepository interface {
	Get(ctx context.Context, userID string) (string, error)
	Set(ctx context.Context, userID string, token string) error
	Delete(ctx context.Context, userID string) error
}
