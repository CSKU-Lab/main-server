package services

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/repositories"
)

// AnalyticsOverview is the full payload for the admin dashboard, assembled from
// several aggregation queries over the given trailing window (days).
type AnalyticsOverview struct {
	Summary           *repositories.AnalyticsSummary  `json:"summary"`
	SubmissionsPerDay []repositories.DailySubmissions `json:"submissions_per_day"`
	ActiveUsersPerDay []repositories.DailyCount       `json:"active_users_per_day"`
	SubmissionsByType []repositories.TypeCount        `json:"submissions_by_type"`
	TopCourses        []repositories.CourseCount      `json:"top_courses"`
}

type AnalyticsService interface {
	GetOverview(ctx context.Context, days int) (*AnalyticsOverview, error)
}

type analyticsService struct {
	repo repositories.AnalyticsRepository
}

func NewAnalyticsService(repo repositories.AnalyticsRepository) AnalyticsService {
	return &analyticsService{repo: repo}
}

const topCoursesLimit = 5

// activeWindowMinutes is the trailing window used to count "currently active"
// users from their last_seen snapshot.
const activeWindowMinutes = 15

func (s *analyticsService) GetOverview(ctx context.Context, days int) (*AnalyticsOverview, error) {
	summary, err := s.repo.GetSummary(ctx, days, activeWindowMinutes)
	if err != nil {
		return nil, err
	}

	submissionsPerDay, err := s.repo.GetSubmissionsPerDay(ctx, days)
	if err != nil {
		return nil, err
	}

	activeUsersPerDay, err := s.repo.GetActiveUsersPerDay(ctx, days)
	if err != nil {
		return nil, err
	}

	submissionsByType, err := s.repo.GetSubmissionsByType(ctx, days)
	if err != nil {
		return nil, err
	}

	topCourses, err := s.repo.GetTopCourses(ctx, days, topCoursesLimit)
	if err != nil {
		return nil, err
	}

	return &AnalyticsOverview{
		Summary:           summary,
		SubmissionsPerDay: submissionsPerDay,
		ActiveUsersPerDay: activeUsersPerDay,
		SubmissionsByType: submissionsByType,
		TopCourses:        topCourses,
	}, nil
}
