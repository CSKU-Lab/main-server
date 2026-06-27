package sqlx

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type analyticsRepository struct {
	db instance
}

func NewAnalyticsRepository(db instance) repositories.AnalyticsRepository {
	return &analyticsRepository{db: db}
}

// created_at columns are stored in UTC (the postgres container runs with the
// default UTC timezone), so day buckets convert UTC -> Asia/Bangkok before
// truncating to a local date.

func (r *analyticsRepository) GetSummary(ctx context.Context, days int) (*repositories.AnalyticsSummary, error) {
	// active_courses is scoped to the selected window ($1) so it tracks the
	// dashboard range selector. Pass rate is all-time over graded submissions.
	query := `
		SELECT
			(SELECT COUNT(*) FROM users) AS total_users,
			(SELECT COUNT(DISTINCT user_id)
			   FROM auth_logs
			  WHERE action = 'sign-in'
			    AND (created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Bangkok')::date
			        = (now() AT TIME ZONE 'Asia/Bangkok')::date) AS active_users_today,
			(SELECT COUNT(*)
			   FROM submissions
			  WHERE (created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Bangkok')::date
			        = (now() AT TIME ZONE 'Asia/Bangkok')::date) AS submissions_today,
			COALESCE((
				SELECT COUNT(*) FILTER (WHERE status = 'passed')::numeric
				       / NULLIF(COUNT(*) FILTER (WHERE status IN ('passed', 'failed')), 0)
				  FROM submissions
			)::float8, 0) AS pass_rate_of_graded,
			(SELECT COUNT(DISTINCT course_id)
			   FROM submissions
			  WHERE created_at >= now() - ($1::int || ' days')::interval) AS active_courses
	`

	var summary repositories.AnalyticsSummary
	if err := r.db.GetContext(ctx, &summary, query, days); err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *analyticsRepository) GetSubmissionsPerDay(ctx context.Context, days int) ([]repositories.DailySubmissions, error) {
	// generate_series gap-fills days with no submissions so the chart x-axis
	// stays continuous and zero days are explicit.
	query := `
		WITH series AS (
			SELECT generate_series(
				(now() AT TIME ZONE 'Asia/Bangkok')::date - ($1::int - 1),
				(now() AT TIME ZONE 'Asia/Bangkok')::date,
				interval '1 day'
			)::date AS d
		)
		SELECT
			to_char(series.d, 'YYYY-MM-DD') AS date,
			COUNT(s.id) AS total,
			COUNT(s.id) FILTER (WHERE s.status = 'passed') AS passed
		FROM series
		LEFT JOIN submissions s
			ON (s.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Bangkok')::date = series.d
		GROUP BY series.d
		ORDER BY series.d
	`

	rows := []repositories.DailySubmissions{}
	if err := r.db.SelectContext(ctx, &rows, query, days); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *analyticsRepository) GetActiveUsersPerDay(ctx context.Context, days int) ([]repositories.DailyCount, error) {
	query := `
		WITH series AS (
			SELECT generate_series(
				(now() AT TIME ZONE 'Asia/Bangkok')::date - ($1::int - 1),
				(now() AT TIME ZONE 'Asia/Bangkok')::date,
				interval '1 day'
			)::date AS d
		)
		SELECT
			to_char(series.d, 'YYYY-MM-DD') AS date,
			COUNT(DISTINCT a.user_id) AS count
		FROM series
		LEFT JOIN auth_logs a
			ON (a.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Bangkok')::date = series.d
			AND a.action = 'sign-in'
		GROUP BY series.d
		ORDER BY series.d
	`

	rows := []repositories.DailyCount{}
	if err := r.db.SelectContext(ctx, &rows, query, days); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *analyticsRepository) GetSubmissionsByType(ctx context.Context, days int) ([]repositories.TypeCount, error) {
	// material_type carries both the legacy 'type' and current 'typing' values;
	// fold them into one bucket so the donut shows a single typing slice.
	query := `
		SELECT
			CASE WHEN m.type = 'type' THEN 'typing' ELSE m.type::text END AS type,
			COUNT(s.id) AS count
		FROM submissions s
		JOIN materials m ON s.material_id = m.id
		WHERE s.created_at >= now() - ($1::int || ' days')::interval
		GROUP BY 1
		ORDER BY count DESC
	`

	rows := []repositories.TypeCount{}
	if err := r.db.SelectContext(ctx, &rows, query, days); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *analyticsRepository) GetTopCourses(ctx context.Context, days int, limit int) ([]repositories.CourseCount, error) {
	query := `
		SELECT
			s.course_id::text AS course_id,
			c.name AS name,
			COUNT(s.id) AS count
		FROM submissions s
		JOIN courses c ON s.course_id = c.id
		WHERE s.created_at >= now() - ($1::int || ' days')::interval
		GROUP BY s.course_id, c.name
		ORDER BY count DESC
		LIMIT $2
	`

	rows := []repositories.CourseCount{}
	if err := r.db.SelectContext(ctx, &rows, query, days, limit); err != nil {
		return nil, err
	}
	return rows, nil
}
