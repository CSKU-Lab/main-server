package repositories

import "context"

// AuthLogRepository appends authentication activity rows into auth_logs.
// Rows drive the analytics daily-active-user aggregation.
type AuthLogRepository interface {
	Create(ctx context.Context, id string, userID string, action string) error
}
