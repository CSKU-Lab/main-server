package permission

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

// service implements the Service interface with repository dependencies.
type service struct {
	userRepo              repositories.User
	courseRepo            repositories.CourseRepository
	courseCreatorRepo     repositories.CourseCreatorRepository
	sectionRepo           repositories.SectionRepository
	sectionInstructorRepo repositories.SectionInstructorRepository
	sectionStudentRepo    repositories.SectionStudentRepository
	submissionRepo        repositories.SubmissionRepository
}

// NewService creates a new permission service with all required repositories.
func NewService(
	userRepo repositories.User,
	courseRepo repositories.CourseRepository,
	courseCreatorRepo repositories.CourseCreatorRepository,
	sectionRepo repositories.SectionRepository,
	sectionInstructorRepo repositories.SectionInstructorRepository,
	sectionStudentRepo repositories.SectionStudentRepository,
	submissionRepo repositories.SubmissionRepository,
) Service {
	return &service{
		userRepo:              userRepo,
		courseRepo:            courseRepo,
		courseCreatorRepo:     courseCreatorRepo,
		sectionRepo:           sectionRepo,
		sectionInstructorRepo: sectionInstructorRepo,
		sectionStudentRepo:    sectionStudentRepo,
		submissionRepo:        submissionRepo,
	}
}

// IsAdmin checks if the user has the admin role.
func (s *service) IsAdmin(ctx context.Context, userID string) (bool, error) {
	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusNotFound,
				Message:    "User not found",
			})
		}
		return false, err
	}

	for _, role := range userData.Roles {
		if role == string(models.ADMIN) {
			return true, nil
		}
	}
	return false, nil
}

// IsInstructor checks if the user has the instructor role.
func (s *service) IsInstructor(ctx context.Context, userID string) (bool, error) {
	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusNotFound,
				Message:    "User not found",
			})
		}
		return false, err
	}

	for _, role := range userData.Roles {
		if role == string(models.INSTRUCTOR) {
			return true, nil
		}
	}
	return false, nil
}

// IsStudent checks if the user has the student role.
func (s *service) IsStudent(ctx context.Context, userID string) (bool, error) {
	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusNotFound,
				Message:    "User not found",
			})
		}
		return false, err
	}

	for _, role := range userData.Roles {
		if role == string(models.STUDENT) {
			return true, nil
		}
	}
	return false, nil
}

// IsCourseCreator checks if the user is a creator of the specified course.
func (s *service) IsCourseCreator(ctx context.Context, userID string, courseID string) (bool, error) {
	creators, err := s.courseCreatorRepo.GetCreators(ctx, courseID)
	if err != nil {
		return false, err
	}

	for _, creator := range creators {
		if creator.ID == userID {
			return true, nil
		}
	}
	return false, nil
}

// IsCourseInstructor checks if the user is an instructor in any section of the course.
func (s *service) IsCourseInstructor(ctx context.Context, userID string, courseID string) (bool, error) {
	// Get all sections for the course
	sections, err := s.sectionRepo.GetByCourseID(ctx, courseID)
	if err != nil {
		return false, err
	}

	// Check if user is an instructor in any section
	for _, section := range sections {
		isInstructor, err := s.IsSectionInstructor(ctx, userID, section.ID)
		if err != nil {
			continue
		}
		if isInstructor {
			return true, nil
		}
	}
	return false, nil
}

// IsSectionInstructor checks if the user is an instructor of the specified section.
func (s *service) IsSectionInstructor(ctx context.Context, userID string, sectionID string) (bool, error) {
	instructors, err := s.sectionInstructorRepo.Get(ctx, sectionID)
	if err != nil {
		return false, err
	}

	for _, instructor := range instructors {
		if instructor.ID == userID {
			return true, nil
		}
	}
	return false, nil
}

// IsSectionStudent checks if the user is a student enrolled in the specified section.
func (s *service) IsSectionStudent(ctx context.Context, userID string, sectionID string) (bool, error) {
	_, err := s.sectionStudentRepo.GetBySectionAndStudentID(ctx, sectionID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetSection retrieves a section by ID.
func (s *service) GetSection(ctx context.Context, sectionID string) (*models.Section, error) {
	sections, err := s.sectionRepo.GetByCourseID(ctx, "")
	if err != nil {
		return nil, err
	}

	// Find the section with matching ID
	for _, section := range sections {
		if section.ID == sectionID {
			return &section, nil
		}
	}

	return nil, cserrors.New(&cserrors.Option{
		HttpStatus: http.StatusNotFound,
		Message:    "Section not found",
	})
}

// IsSubmissionOwner checks if the user owns the specified submission.
func (s *service) IsSubmissionOwner(ctx context.Context, userID string, submissionID string) (bool, error) {
	submission, err := s.submissionRepo.Get(ctx, submissionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusNotFound,
				Message:    "Submission not found",
			})
		}
		return false, err
	}

	return submission.UserID == userID, nil
}

// GetSubmission retrieves a submission by ID.
func (s *service) GetSubmission(ctx context.Context, submissionID string) (*repositories.Submission, error) {
	return s.submissionRepo.Get(ctx, submissionID)
}
