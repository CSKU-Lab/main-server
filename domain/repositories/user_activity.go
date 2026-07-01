package repositories

import "context"

// UserActivityRepository keeps a per-user last_seen snapshot, upserted on every
// authenticated request. It backs the real-time "currently active" count, which
// auth_logs alone cannot provide since long sessions rarely re-authenticate.
type UserActivityRepository interface {
	Touch(ctx context.Context, userID string) error
}
