package repositories

import "context"

type CourseEnrollmentRepository interface {
	Enroll(ctx context.Context, courseID string, studentID string) error
	Unenroll(ctx context.Context, courseID string, studentID string) error
	IsEnrolled(ctx context.Context, courseID string, studentID string) (bool, error)
	CountByCourseID(ctx context.Context, courseID string) (int, error)
}
