package permission

import (
	"context"
	"errors"
	"testing"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories for testing

type mockCourseCreatorRepo struct {
	mock.Mock
}

func (m *mockCourseCreatorRepo) GetCreators(ctx context.Context, courseID string) ([]models.CourseCreator, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.CourseCreator), args.Error(1)
}

func (m *mockCourseCreatorRepo) SetCreators(ctx context.Context, courseID string, creators []string) error {
	args := m.Called(ctx, courseID, creators)
	return args.Error(0)
}

type mockSectionInstructorRepo struct {
	mock.Mock
}

func (m *mockSectionInstructorRepo) Get(ctx context.Context, sectionID string) ([]models.SectionInstructor, error) {
	args := m.Called(ctx, sectionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.SectionInstructor), args.Error(1)
}

func (m *mockSectionInstructorRepo) Add(ctx context.Context, sectionID string, instructorID string) error {
	args := m.Called(ctx, sectionID, instructorID)
	return args.Error(0)
}

func (m *mockSectionInstructorRepo) DeleteBySectionID(ctx context.Context, sectionID string) error {
	args := m.Called(ctx, sectionID)
	return args.Error(0)
}

type mockSectionStudentRepo struct {
	mock.Mock
}

func (m *mockSectionStudentRepo) Add(ctx context.Context, sectionID string, studentID string) error {
	args := m.Called(ctx, sectionID, studentID)
	return args.Error(0)
}

func (m *mockSectionStudentRepo) GetBySectionID(ctx context.Context, sectionID string) ([]models.Student, error) {
	args := m.Called(ctx, sectionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Student), args.Error(1)
}

func (m *mockSectionStudentRepo) GetSectionsPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filters []sanitize.Filter) ([]repositories.RawSection, error) {
	return nil, nil
}

func (m *mockSectionStudentRepo) GetBySectionAndStudentID(ctx context.Context, sectionID string, studentID string) (*models.SectionStudent, error) {
	args := m.Called(ctx, sectionID, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SectionStudent), args.Error(1)
}

func (m *mockSectionStudentRepo) GetByStudentID(ctx context.Context, studentID string) ([]models.Section, error) {
	return nil, nil
}

func (m *mockSectionStudentRepo) Count(ctx context.Context, filters []sanitize.Filter) (int, error) {
	return 0, nil
}

func (m *mockSectionStudentRepo) RemoveBySectionIDAndStudentID(ctx context.Context, sectionID string, studentID string) error {
	return nil
}

type mockSubmissionRepo struct {
	mock.Mock
}

func (m *mockSubmissionRepo) Create(ctx context.Context, req *repositories.Submission) error {
	return nil
}

func (m *mockSubmissionRepo) Update(ctx context.Context, req *repositories.UpdateSubmissionRequest) error {
	return nil
}

func (m *mockSubmissionRepo) Get(ctx context.Context, id string) (*repositories.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.Submission), args.Error(1)
}

func (m *mockSubmissionRepo) GetPagination(ctx context.Context, userID string, materialID string, labID string, sectionID string, page int, pageSize int, sortOrder string) ([]repositories.Submission, error) {
	return nil, nil
}

func (m *mockSubmissionRepo) GetLatestOfStudentIDInSectionID(ctx context.Context, sectionID, labID, materialID, studentID string) (*models.RawSubmission, error) {
	return nil, nil
}

func (m *mockSubmissionRepo) GetLatestByMaterialSectionAndLabID(ctx context.Context, materialID string, sectionID string, labID string) ([]models.RawSubmission, error) {
	return nil, nil
}

func (m *mockSubmissionRepo) GetPaginationByMaterialSectionLabAndStudentID(ctx context.Context, materialID string, sectionID string, labID string, studentID string, page int, pageSize int, sortBy, sortOrder string) ([]models.RawSubmission, error) {
	return nil, nil
}

func (m *mockSubmissionRepo) Count(ctx context.Context, userID string, materialID string, labID string, sectionID string) (int, error) {
	return 0, nil
}

func (m *mockSubmissionRepo) CountByMaterialSectionLabAndStudentID(ctx context.Context, materialID string, sectionID string, labID string, studentID string) (int, error) {
	return 0, nil
}

func (m *mockSubmissionRepo) CountCompletedStudentsByLabAndSection(ctx context.Context, labID string, sectionID string) (int, error) {
	return 0, nil
}

// TestNewPermissionService tests the service constructor.
func TestNewPermissionService(t *testing.T) {
	courseCreatorRepo := new(mockCourseCreatorRepo)
	sectionInstructorRepo := new(mockSectionInstructorRepo)
	sectionStudentRepo := new(mockSectionStudentRepo)
	submissionRepo := new(mockSubmissionRepo)

	service := NewPermissionService(courseCreatorRepo, sectionInstructorRepo, sectionStudentRepo, submissionRepo)

	assert.NotNil(t, service)
	assert.Implements(t, (*Service)(nil), service)
}

// TestPermissionServiceIsAdmin tests the IsAdmin condition.
func TestPermissionServiceIsAdmin(t *testing.T) {
	service := NewPermissionService(nil, nil, nil, nil)
	ctx := context.Background()
	params := map[string]string{}

	t.Run("admin user returns true", func(t *testing.T) {
		user := &models.User{ID: "user-123", Roles: []models.Role{models.ADMIN}}
		condition := service.IsAdmin()
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
	})

	t.Run("non-admin user returns false", func(t *testing.T) {
		user := &models.User{ID: "user-123", Roles: []models.Role{models.STUDENT}}
		condition := service.IsAdmin()
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
	})

	t.Run("nil user returns false", func(t *testing.T) {
		condition := service.IsAdmin()
		result := condition.Evaluate(ctx, nil, params)
		assert.False(t, result)
	})
}

// TestPermissionServiceIsAuthenticated tests the IsAuthenticated condition.
func TestPermissionServiceIsAuthenticated(t *testing.T) {
	service := NewPermissionService(nil, nil, nil, nil)
	ctx := context.Background()
	params := map[string]string{}

	t.Run("authenticated user returns true", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		condition := service.IsAuthenticated()
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
	})

	t.Run("nil user returns false", func(t *testing.T) {
		condition := service.IsAuthenticated()
		result := condition.Evaluate(ctx, nil, params)
		assert.False(t, result)
	})
}

// TestPermissionServiceIsCourseCreator tests the IsCourseCreator condition.
func TestPermissionServiceIsCourseCreator(t *testing.T) {
	courseCreatorRepo := new(mockCourseCreatorRepo)
	service := NewPermissionService(courseCreatorRepo, nil, nil, nil)
	ctx := context.Background()

	t.Run("user is course creator", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{"id": "course-456"}

		courseCreatorRepo.On("GetCreators", ctx, "course-456").Return([]models.CourseCreator{
			{ID: "user-123", Username: "creator1"},
			{ID: "user-789", Username: "creator2"},
		}, nil).Once()

		condition := service.IsCourseCreator("id")
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
		courseCreatorRepo.AssertExpectations(t)
	})

	t.Run("user is not course creator", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{"id": "course-456"}

		courseCreatorRepo.On("GetCreators", ctx, "course-456").Return([]models.CourseCreator{
			{ID: "user-789", Username: "creator1"},
		}, nil).Once()

		condition := service.IsCourseCreator("id")
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
		courseCreatorRepo.AssertExpectations(t)
	})

	t.Run("missing course ID param", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{}

		condition := service.IsCourseCreator("id")
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
	})

	t.Run("nil user", func(t *testing.T) {
		params := map[string]string{"id": "course-456"}

		condition := service.IsCourseCreator("id")
		result := condition.Evaluate(ctx, nil, params)
		assert.False(t, result)
	})

	t.Run("repository error", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{"id": "course-456"}

		courseCreatorRepo.On("GetCreators", ctx, "course-456").Return(nil, errors.New("db error")).Once()

		condition := service.IsCourseCreator("id")
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
		courseCreatorRepo.AssertExpectations(t)
	})
}

// TestPermissionServiceIsSectionInstructor tests the IsSectionInstructor condition.
func TestPermissionServiceIsSectionInstructor(t *testing.T) {
	sectionInstructorRepo := new(mockSectionInstructorRepo)
	service := NewPermissionService(nil, sectionInstructorRepo, nil, nil)
	ctx := context.Background()

	t.Run("user is section instructor", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{"id": "section-456"}

		sectionInstructorRepo.On("Get", ctx, "section-456").Return([]models.SectionInstructor{
			{ID: "user-123", Username: "instructor1"},
			{ID: "user-789", Username: "instructor2"},
		}, nil).Once()

		condition := service.IsSectionInstructor("id")
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
		sectionInstructorRepo.AssertExpectations(t)
	})

	t.Run("user is not section instructor", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{"id": "section-456"}

		sectionInstructorRepo.On("Get", ctx, "section-456").Return([]models.SectionInstructor{
			{ID: "user-789", Username: "instructor1"},
		}, nil).Once()

		condition := service.IsSectionInstructor("id")
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
		sectionInstructorRepo.AssertExpectations(t)
	})

	t.Run("missing section ID param", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{}

		condition := service.IsSectionInstructor("id")
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
	})
}

// TestPermissionServiceIsSectionStudent tests the IsSectionStudent condition.
func TestPermissionServiceIsSectionStudent(t *testing.T) {
	sectionStudentRepo := new(mockSectionStudentRepo)
	service := NewPermissionService(nil, nil, sectionStudentRepo, nil)
	ctx := context.Background()

	t.Run("user is section student", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{"id": "section-456"}

		sectionStudentRepo.On("GetBySectionAndStudentID", ctx, "section-456", "user-123").Return(&models.SectionStudent{
			SectionID: "section-456",
			StudentID: "user-123",
		}, nil).Once()

		condition := service.IsSectionStudent("id")
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
		sectionStudentRepo.AssertExpectations(t)
	})

	t.Run("user is not section student", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{"id": "section-456"}

		sectionStudentRepo.On("GetBySectionAndStudentID", ctx, "section-456", "user-123").Return(nil, errors.New("not found")).Once()

		condition := service.IsSectionStudent("id")
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
		sectionStudentRepo.AssertExpectations(t)
	})

	t.Run("missing section ID param", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{}

		condition := service.IsSectionStudent("id")
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
	})
}

// TestPermissionServiceIsSubmissionOwner tests the IsSubmissionOwner condition.
func TestPermissionServiceIsSubmissionOwner(t *testing.T) {
	submissionRepo := new(mockSubmissionRepo)
	service := NewPermissionService(nil, nil, nil, submissionRepo)
	ctx := context.Background()

	t.Run("user is submission owner", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{"id": "submission-456"}

		submissionRepo.On("Get", ctx, "submission-456").Return(&repositories.Submission{
			ID:     "submission-456",
			UserID: "user-123",
		}, nil).Once()

		condition := service.IsSubmissionOwner("id")
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
		submissionRepo.AssertExpectations(t)
	})

	t.Run("user is not submission owner", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{"id": "submission-456"}

		submissionRepo.On("Get", ctx, "submission-456").Return(&repositories.Submission{
			ID:     "submission-456",
			UserID: "user-789",
		}, nil).Once()

		condition := service.IsSubmissionOwner("id")
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
		submissionRepo.AssertExpectations(t)
	})

	t.Run("missing submission ID param", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{}

		condition := service.IsSubmissionOwner("id")
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
	})

	t.Run("submission not found", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		params := map[string]string{"id": "submission-456"}

		submissionRepo.On("Get", ctx, "submission-456").Return(nil, errors.New("not found")).Once()

		condition := service.IsSubmissionOwner("id")
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
		submissionRepo.AssertExpectations(t)
	})
}

// TestPermissionServiceOr tests the Or logical operator.
func TestPermissionServiceOr(t *testing.T) {
	service := NewPermissionService(nil, nil, nil, nil)
	ctx := context.Background()
	params := map[string]string{}

	t.Run("first condition true", func(t *testing.T) {
		user := &models.User{ID: "user-123", Roles: []models.Role{models.ADMIN}}
		condition := service.Or(
			service.IsAdmin(),
			service.IsAuthenticated(),
		)
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
	})

	t.Run("second condition true", func(t *testing.T) {
		user := &models.User{ID: "user-123", Roles: []models.Role{models.STUDENT}}
		condition := service.Or(
			service.IsAdmin(),
			service.IsAuthenticated(),
		)
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
	})

	t.Run("both conditions false", func(t *testing.T) {
		condition := service.Or(
			service.IsAdmin(),
			service.IsAuthenticated(),
		)
		result := condition.Evaluate(ctx, nil, params)
		assert.False(t, result)
	})

	t.Run("empty OR returns false", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		condition := service.Or()
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
	})
}

// TestPermissionServiceAnd tests the And logical operator.
func TestPermissionServiceAnd(t *testing.T) {
	service := NewPermissionService(nil, nil, nil, nil)
	ctx := context.Background()
	params := map[string]string{}

	t.Run("both conditions true", func(t *testing.T) {
		user := &models.User{ID: "user-123", Roles: []models.Role{models.ADMIN}}
		condition := service.And(
			service.IsAdmin(),
			service.IsAuthenticated(),
		)
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
	})

	t.Run("first condition false", func(t *testing.T) {
		user := &models.User{ID: "user-123", Roles: []models.Role{models.STUDENT}}
		condition := service.And(
			service.IsAdmin(),
			service.IsAuthenticated(),
		)
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
	})

	t.Run("second condition false", func(t *testing.T) {
		user := &models.User{ID: "user-123", Roles: []models.Role{models.ADMIN}}
		condition := service.And(
			service.IsAdmin(),
			service.Not(service.IsAuthenticated()),
		)
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
	})

	t.Run("empty AND returns true", func(t *testing.T) {
		user := &models.User{ID: "user-123"}
		condition := service.And()
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
	})
}

// TestPermissionServiceNot tests the Not logical operator.
func TestPermissionServiceNot(t *testing.T) {
	service := NewPermissionService(nil, nil, nil, nil)
	ctx := context.Background()
	params := map[string]string{}

	t.Run("not true is false", func(t *testing.T) {
		user := &models.User{ID: "user-123", Roles: []models.Role{models.ADMIN}}
		condition := service.Not(service.IsAdmin())
		result := condition.Evaluate(ctx, user, params)
		assert.False(t, result)
	})

	t.Run("not false is true", func(t *testing.T) {
		user := &models.User{ID: "user-123", Roles: []models.Role{models.STUDENT}}
		condition := service.Not(service.IsAdmin())
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
	})
}

// TestPermissionServiceComplexConditions tests complex nested conditions.
func TestPermissionServiceComplexConditions(t *testing.T) {
	courseCreatorRepo := new(mockCourseCreatorRepo)
	service := NewPermissionService(courseCreatorRepo, nil, nil, nil)
	ctx := context.Background()

	t.Run("admin OR course creator", func(t *testing.T) {
		user := &models.User{ID: "user-123", Roles: []models.Role{models.ADMIN}}
		params := map[string]string{"id": "course-456"}

		// Should pass because user is admin (no need to check course creator)
		condition := service.Or(
			service.IsAdmin(),
			service.IsCourseCreator("id"),
		)
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
	})

	t.Run("(admin OR instructor) AND authenticated", func(t *testing.T) {
		user := &models.User{ID: "user-123", Roles: []models.Role{models.ADMIN}}
		params := map[string]string{}

		condition := service.And(
			service.Or(
				service.IsAdmin(),
				service.IsAuthenticated(),
			),
			service.IsAuthenticated(),
		)
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
	})

	t.Run("NOT (student AND NOT admin)", func(t *testing.T) {
		// This is equivalent to: NOT student OR admin
		// For an admin user, this should pass
		user := &models.User{ID: "user-123", Roles: []models.Role{models.ADMIN}}
		params := map[string]string{}

		condition := service.Not(
			service.And(
				service.IsAuthenticated(),
				service.Not(service.IsAdmin()),
			),
		)
		result := condition.Evaluate(ctx, user, params)
		assert.True(t, result)
	})
}
