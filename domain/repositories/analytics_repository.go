package repositories

import "context"

// AnalyticsSummary holds the headline numbers shown in the dashboard cards.
type AnalyticsSummary struct {
	TotalUsers           int     `db:"total_users" json:"total_users"`
	ActiveUsersToday     int     `db:"active_users_today" json:"active_users_today"`
	CurrentlyActiveUsers int     `db:"currently_active_users" json:"currently_active_users"`
	SubmissionsToday     int     `db:"submissions_today" json:"submissions_today"`
	PassRateOfGraded     float64 `db:"pass_rate_of_graded" json:"pass_rate_of_graded"`
	ActiveCourses        int     `db:"active_courses" json:"active_courses"`
}

// DailySubmissions is one day's submission counts in the trend chart.
type DailySubmissions struct {
	Date   string `db:"date" json:"date"`
	Total  int    `db:"total" json:"total"`
	Passed int    `db:"passed" json:"passed"`
}

// DailyCount is one day's distinct-user count in the activity trend.
type DailyCount struct {
	Date  string `db:"date" json:"date"`
	Count int    `db:"count" json:"count"`
}

// TypeCount is the submission count for one material type.
type TypeCount struct {
	Type  string `db:"type" json:"type"`
	Count int    `db:"count" json:"count"`
}

// CourseCount is the submission count for one course.
type CourseCount struct {
	CourseID string `db:"course_id" json:"course_id"`
	Name     string `db:"name" json:"name"`
	Count    int    `db:"count" json:"count"`
}

// AnalyticsRepository exposes read-only aggregation queries over existing
// domain tables. All day buckets are computed in the Asia/Bangkok timezone.
type AnalyticsRepository interface {
	GetSummary(ctx context.Context, days int, activeWindowMinutes int) (*AnalyticsSummary, error)
	GetSubmissionsPerDay(ctx context.Context, days int) ([]DailySubmissions, error)
	GetActiveUsersPerDay(ctx context.Context, days int) ([]DailyCount, error)
	GetSubmissionsByType(ctx context.Context, days int) ([]TypeCount, error)
	GetTopCourses(ctx context.Context, days int, limit int) ([]CourseCount, error)
}
