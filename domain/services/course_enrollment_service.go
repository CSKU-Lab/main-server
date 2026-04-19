package services

import (
	"context"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type CourseEnrollmentService interface {
	Enroll(ctx context.Context, courseID string, studentID string) error
	Unenroll(ctx context.Context, courseID string, studentID string) error
	IsEnrolled(ctx context.Context, courseID string, studentID string) (bool, error)
}

type courseEnrollmentService struct {
	enrollmentRepo repositories.CourseEnrollmentRepository
	courseRepo     repositories.CourseRepository
}

func NewCourseEnrollmentService(enrollmentRepo repositories.CourseEnrollmentRepository, courseRepo repositories.CourseRepository) CourseEnrollmentService {
	return &courseEnrollmentService{
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
	}
}

func (s *courseEnrollmentService) Enroll(ctx context.Context, courseID string, studentID string) error {
	course, err := s.courseRepo.GetByID(ctx, courseID)
	if err != nil {
		return err
	}

	if course.Visibility != "public" {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Message:    "Course is not public",
		})
	}

	return s.enrollmentRepo.Enroll(ctx, courseID, studentID)
}

func (s *courseEnrollmentService) Unenroll(ctx context.Context, courseID string, studentID string) error {
	return s.enrollmentRepo.Unenroll(ctx, courseID, studentID)
}

func (s *courseEnrollmentService) IsEnrolled(ctx context.Context, courseID string, studentID string) (bool, error) {
	return s.enrollmentRepo.IsEnrolled(ctx, courseID, studentID)
}
