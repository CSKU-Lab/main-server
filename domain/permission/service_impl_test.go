package permission

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repositories
type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*repositories.UserData, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.UserData), args.Error(1)
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*repositories.UserData, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.UserData), args.Error(1)
}

func (m *mockUserRepo) GetManyByUsername(ctx context.Context, usernames []string) ([]repositories.UserData, error) {
	args := m.Called(ctx, usernames)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.UserData), args.Error(1)
}

func (m *mockUserRepo) GetManyByFindBy(ctx context.Context, data []string, findBy string, role string) ([]repositories.UserData, error) {
	args := m.Called(ctx, data, findBy, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.UserData), args.Error(1)
}

func (m *mockUserRepo) GetByID(ctx context.Context, ID string) (*repositories.UserData, error) {
	args := m.Called(ctx, ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.UserData), args.Error(1)
}

func (m *mockUserRepo) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter) ([]repositories.UserData, error) {
	args := m.Called(ctx, page, limit, search, sortBy, sortOrder, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.UserData), args.Error(1)
}

func (m *mockUserRepo) Count(ctx context.Context, search string, filters []sanitize.Filter) (int, error) {
	args := m.Called(ctx, search, filters)
	return args.Int(0), args.Error(1)
}

func (m *mockUserRepo) Create(ctx context.Context, user repositories.CreateMultiTypeUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) Update(ctx context.Context, ID string, user *requests.UpdateUser) error {
	args := m.Called(ctx, ID, user)
	return args.Error(0)
}

func (m *mockUserRepo) Delete(ctx context.Context, ID string) error {
	args := m.Called(ctx, ID)
	return args.Error(0)
}

func (m *mockUserRepo) DeleteMany(ctx context.Context, IDs []string) error {
	args := m.Called(ctx, IDs)
	return args.Error(0)
}

type mockCourseRepo struct {
	mock.Mock
}

func (m *mockCourseRepo) Create(ctx context.Context, ID string, c *requests.CreateCourse) error {
	args := m.Called(ctx, ID, c)
	return args.Error(0)
}

func (m *mockCourseRepo) GetByID(ctx context.Context, ID string) (*models.Course, error) {
	args := m.Called(ctx, ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *mockCourseRepo) GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string, show string) ([]models.Course, error) {
	args := m.Called(ctx, page, pageSize, search, sortBy, sortOrder, show)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *mockCourseRepo) Count(ctx context.Context, search string, show string) (int, error) {
	args := m.Called(ctx, search, show)
	return args.Int(0), args.Error(1)
}

func (m *mockCourseRepo) UpdateByID(ctx context.Context, ID string, c *requests.UpdateCourse) error {
	args := m.Called(ctx, ID, c)
	return args.Error(0)
}

func (m *mockCourseRepo) DeleteByID(ctx context.Context, ID string) error {
	args := m.Called(ctx, ID)
	return args.Error(0)
}

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

type mockSectionRepo struct {
	mock.Mock
}

func (m *mockSectionRepo) Create(ctx context.Context, ID string, section *repositories.CreateSection) error {
	args := m.Called(ctx, ID, section)
	return args.Error(0)
}

func (m *mockSectionRepo) UpdateByID(ctx context.Context, ID string, section *repositories.UpdateSection) error {
	args := m.Called(ctx, ID, section)
	return args.Error(0)
}

func (m *mockSectionRepo) GetByID(ctx context.Context, ID string) (*repositories.RawSection, error) {
	args := m.Called(ctx, ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.RawSection), args.Error(1)
}

func (m *mockSectionRepo) GetBySemesterID(ctx context.Context, semesterID string) ([]models.Section, error) {
	args := m.Called(ctx, semesterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Section), args.Error(1)
}

func (m *mockSectionRepo) GetByCourseID(ctx context.Context, courseID string) ([]models.Section, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Section), args.Error(1)
}

func (m *mockSectionRepo) GetByCourseIDAndSemesterID(ctx context.Context, courseID string, semesterID string) ([]models.Section, error) {
	args := m.Called(ctx, courseID, semesterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Section), args.Error(1)
}

func (m *mockSectionRepo) GetRawBySemesterID(ctx context.Context, semesterID string) ([]repositories.RawSection, error) {
	args := m.Called(ctx, semesterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.RawSection), args.Error(1)
}

func (m *mockSectionRepo) DeleteByID(ctx context.Context, ID string) error {
	args := m.Called(ctx, ID)
	return args.Error(0)
}

func (m *mockSectionRepo) DeleteByCourseID(ctx context.Context, courseID string) error {
	args := m.Called(ctx, courseID)
	return args.Error(0)
}

func (m *mockSectionRepo) GetRawByID(ctx context.Context, ID string) (*repositories.RawSection, error) {
	args := m.Called(ctx, ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.RawSection), args.Error(1)
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

func (m *mockSectionInstructorRepo) Add(ctx context.Context, sectionID string, id string) error {
	args := m.Called(ctx, sectionID, id)
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
	args := m.Called(ctx, page, limit, sortBy, sortOrder, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.RawSection), args.Error(1)
}

func (m *mockSectionStudentRepo) GetBySectionAndStudentID(ctx context.Context, sectionID string, studentID string) (*models.SectionStudent, error) {
	args := m.Called(ctx, sectionID, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SectionStudent), args.Error(1)
}

func (m *mockSectionStudentRepo) GetByStudentID(ctx context.Context, studentID string) ([]models.Section, error) {
	args := m.Called(ctx, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Section), args.Error(1)
}

func (m *mockSectionStudentRepo) Count(ctx context.Context, filters []sanitize.Filter) (int, error) {
	args := m.Called(ctx, filters)
	return args.Int(0), args.Error(1)
}

func (m *mockSectionStudentRepo) RemoveBySectionIDAndStudentID(ctx context.Context, sectionID string, studentID string) error {
	args := m.Called(ctx, sectionID, studentID)
	return args.Error(0)
}

type mockSubmissionRepo struct {
	mock.Mock
}

func (m *mockSubmissionRepo) Create(ctx context.Context, req *repositories.Submission) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockSubmissionRepo) Update(ctx context.Context, req *repositories.UpdateSubmissionRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockSubmissionRepo) Get(ctx context.Context, id string) (*repositories.Submission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.Submission), args.Error(1)
}

func (m *mockSubmissionRepo) GetPagination(ctx context.Context, userID string, materialID string, labID string, sectionID string, page int, pageSize int, sortOrder string) ([]repositories.Submission, error) {
	args := m.Called(ctx, userID, materialID, labID, sectionID, page, pageSize, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.Submission), args.Error(1)
}

func (m *mockSubmissionRepo) GetLatestOfStudentIDInSectionID(ctx context.Context, sectionID, labID, materialID, studentID string) (*models.RawSubmission, error) {
	args := m.Called(ctx, sectionID, labID, materialID, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RawSubmission), args.Error(1)
}

func (m *mockSubmissionRepo) GetLatestByMaterialSectionAndLabID(ctx context.Context, materialID string, sectionID string, labID string) ([]models.RawSubmission, error) {
	args := m.Called(ctx, materialID, sectionID, labID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RawSubmission), args.Error(1)
}

func (m *mockSubmissionRepo) GetPaginationByMaterialSectionLabAndStudentID(ctx context.Context, materialID string, sectionID string, labID string, studentID string, page int, pageSize int, sortBy, sortOrder string) ([]models.RawSubmission, error) {
	args := m.Called(ctx, materialID, sectionID, labID, studentID, page, pageSize, sortBy, sortOrder)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RawSubmission), args.Error(1)
}

func (m *mockSubmissionRepo) Count(ctx context.Context, userID string, materialID string, labID string, sectionID string) (int, error) {
	args := m.Called(ctx, userID, materialID, labID, sectionID)
	return args.Int(0), args.Error(1)
}

func (m *mockSubmissionRepo) CountByMaterialSectionLabAndStudentID(ctx context.Context, materialID string, sectionID string, labID string, studentID string) (int, error) {
	args := m.Called(ctx, materialID, sectionID, labID, studentID)
	return args.Int(0), args.Error(1)
}

func (m *mockSubmissionRepo) CountCompletedStudentsByLabAndSection(ctx context.Context, labID string, sectionID string) (int, error) {
	args := m.Called(ctx, labID, sectionID)
	return args.Int(0), args.Error(1)
}

// Helper function to create a new permission service with mocks
func setupTestService() (*service, *mockUserRepo, *mockCourseRepo, *mockCourseCreatorRepo, *mockSectionRepo, *mockSectionInstructorRepo, *mockSectionStudentRepo, *mockSubmissionRepo) {
	userRepo := new(mockUserRepo)
	courseRepo := new(mockCourseRepo)
	courseCreatorRepo := new(mockCourseCreatorRepo)
	sectionRepo := new(mockSectionRepo)
	sectionInstructorRepo := new(mockSectionInstructorRepo)
	sectionStudentRepo := new(mockSectionStudentRepo)
	submissionRepo := new(mockSubmissionRepo)

	svc := NewService(
		userRepo,
		courseRepo,
		courseCreatorRepo,
		sectionRepo,
		sectionInstructorRepo,
		sectionStudentRepo,
		submissionRepo,
	).(*service)

	return svc, userRepo, courseRepo, courseCreatorRepo, sectionRepo, sectionInstructorRepo, sectionStudentRepo, submissionRepo
}

// Test IsAdmin
func TestIsAdmin_Success(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _ := setupTestService()
	ctx := context.Background()

	userData := &repositories.UserData{
		ID:    "user-123",
		Roles: []string{"admin"},
	}

	userRepo.On("GetByID", ctx, "user-123").Return(userData, nil)

	result, err := svc.IsAdmin(ctx, "user-123")

	assert.NoError(t, err)
	assert.True(t, result)
	userRepo.AssertExpectations(t)
}

func TestIsAdmin_NotAdmin(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _ := setupTestService()
	ctx := context.Background()

	userData := &repositories.UserData{
		ID:    "user-123",
		Roles: []string{"student"},
	}

	userRepo.On("GetByID", ctx, "user-123").Return(userData, nil)

	result, err := svc.IsAdmin(ctx, "user-123")

	assert.NoError(t, err)
	assert.False(t, result)
	userRepo.AssertExpectations(t)
}

func TestIsAdmin_UserNotFound(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _ := setupTestService()
	ctx := context.Background()

	userRepo.On("GetByID", ctx, "user-123").Return(nil, sql.ErrNoRows)

	result, err := svc.IsAdmin(ctx, "user-123")

	assert.Error(t, err)
	assert.False(t, result)

	var csErr *cserrors.Error
	assert.True(t, errors.As(err, &csErr))
	assert.Equal(t, http.StatusNotFound, csErr.HttpStatus)
	userRepo.AssertExpectations(t)
}

func TestIsAdmin_RepositoryError(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _ := setupTestService()
	ctx := context.Background()

	userRepo.On("GetByID", ctx, "user-123").Return(nil, errors.New("db error"))

	result, err := svc.IsAdmin(ctx, "user-123")

	assert.Error(t, err)
	assert.False(t, result)
	userRepo.AssertExpectations(t)
}

// Test IsInstructor
func TestIsInstructor_Success(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _ := setupTestService()
	ctx := context.Background()

	userData := &repositories.UserData{
		ID:    "user-123",
		Roles: []string{"instructor"},
	}

	userRepo.On("GetByID", ctx, "user-123").Return(userData, nil)

	result, err := svc.IsInstructor(ctx, "user-123")

	assert.NoError(t, err)
	assert.True(t, result)
	userRepo.AssertExpectations(t)
}

func TestIsInstructor_NotInstructor(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _ := setupTestService()
	ctx := context.Background()

	userData := &repositories.UserData{
		ID:    "user-123",
		Roles: []string{"student"},
	}

	userRepo.On("GetByID", ctx, "user-123").Return(userData, nil)

	result, err := svc.IsInstructor(ctx, "user-123")

	assert.NoError(t, err)
	assert.False(t, result)
	userRepo.AssertExpectations(t)
}

func TestIsInstructor_UserNotFound(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _ := setupTestService()
	ctx := context.Background()

	userRepo.On("GetByID", ctx, "user-123").Return(nil, sql.ErrNoRows)

	result, err := svc.IsInstructor(ctx, "user-123")

	assert.Error(t, err)
	assert.False(t, result)
	userRepo.AssertExpectations(t)
}

// Test IsStudent
func TestIsStudent_Success(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _ := setupTestService()
	ctx := context.Background()

	userData := &repositories.UserData{
		ID:    "user-123",
		Roles: []string{"student"},
	}

	userRepo.On("GetByID", ctx, "user-123").Return(userData, nil)

	result, err := svc.IsStudent(ctx, "user-123")

	assert.NoError(t, err)
	assert.True(t, result)
	userRepo.AssertExpectations(t)
}

func TestIsStudent_NotStudent(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _ := setupTestService()
	ctx := context.Background()

	userData := &repositories.UserData{
		ID:    "user-123",
		Roles: []string{"admin"},
	}

	userRepo.On("GetByID", ctx, "user-123").Return(userData, nil)

	result, err := svc.IsStudent(ctx, "user-123")

	assert.NoError(t, err)
	assert.False(t, result)
	userRepo.AssertExpectations(t)
}

func TestIsStudent_UserNotFound(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _ := setupTestService()
	ctx := context.Background()

	userRepo.On("GetByID", ctx, "user-123").Return(nil, sql.ErrNoRows)

	result, err := svc.IsStudent(ctx, "user-123")

	assert.Error(t, err)
	assert.False(t, result)
	userRepo.AssertExpectations(t)
}

// Test IsCourseCreator
func TestIsCourseCreator_Success(t *testing.T) {
	svc, _, _, courseCreatorRepo, _, _, _, _ := setupTestService()
	ctx := context.Background()

	creators := []models.CourseCreator{
		{ID: "user-123", Username: "creator1"},
		{ID: "user-456", Username: "creator2"},
	}

	courseCreatorRepo.On("GetCreators", ctx, "course-123").Return(creators, nil)

	result, err := svc.IsCourseCreator(ctx, "user-123", "course-123")

	assert.NoError(t, err)
	assert.True(t, result)
	courseCreatorRepo.AssertExpectations(t)
}

func TestIsCourseCreator_NotCreator(t *testing.T) {
	svc, _, _, courseCreatorRepo, _, _, _, _ := setupTestService()
	ctx := context.Background()

	creators := []models.CourseCreator{
		{ID: "user-456", Username: "creator1"},
	}

	courseCreatorRepo.On("GetCreators", ctx, "course-123").Return(creators, nil)

	result, err := svc.IsCourseCreator(ctx, "user-123", "course-123")

	assert.NoError(t, err)
	assert.False(t, result)
	courseCreatorRepo.AssertExpectations(t)
}

func TestIsCourseCreator_EmptyCreators(t *testing.T) {
	svc, _, _, courseCreatorRepo, _, _, _, _ := setupTestService()
	ctx := context.Background()

	creators := []models.CourseCreator{}

	courseCreatorRepo.On("GetCreators", ctx, "course-123").Return(creators, nil)

	result, err := svc.IsCourseCreator(ctx, "user-123", "course-123")

	assert.NoError(t, err)
	assert.False(t, result)
	courseCreatorRepo.AssertExpectations(t)
}

func TestIsCourseCreator_RepositoryError(t *testing.T) {
	svc, _, _, courseCreatorRepo, _, _, _, _ := setupTestService()
	ctx := context.Background()

	courseCreatorRepo.On("GetCreators", ctx, "course-123").Return(nil, errors.New("db error"))

	result, err := svc.IsCourseCreator(ctx, "user-123", "course-123")

	assert.Error(t, err)
	assert.False(t, result)
	courseCreatorRepo.AssertExpectations(t)
}

// Test IsCourseInstructor
func TestIsCourseInstructor_Success(t *testing.T) {
	svc, _, _, _, sectionRepo, sectionInstructorRepo, _, _ := setupTestService()
	ctx := context.Background()

	sections := []models.Section{
		{ID: "section-1", CourseID: "course-123"},
		{ID: "section-2", CourseID: "course-123"},
	}

	instructors := []models.SectionInstructor{
		{ID: "user-123", Username: "instructor1"},
	}

	sectionRepo.On("GetByCourseID", ctx, "course-123").Return(sections, nil)
	sectionInstructorRepo.On("Get", ctx, "section-1").Return([]models.SectionInstructor{}, nil)
	sectionInstructorRepo.On("Get", ctx, "section-2").Return(instructors, nil)

	result, err := svc.IsCourseInstructor(ctx, "user-123", "course-123")

	assert.NoError(t, err)
	assert.True(t, result)
	sectionRepo.AssertExpectations(t)
	sectionInstructorRepo.AssertExpectations(t)
}

func TestIsCourseInstructor_NotInstructor(t *testing.T) {
	svc, _, _, _, sectionRepo, sectionInstructorRepo, _, _ := setupTestService()
	ctx := context.Background()

	sections := []models.Section{
		{ID: "section-1", CourseID: "course-123"},
	}

	instructors := []models.SectionInstructor{
		{ID: "user-456", Username: "instructor1"},
	}

	sectionRepo.On("GetByCourseID", ctx, "course-123").Return(sections, nil)
	sectionInstructorRepo.On("Get", ctx, "section-1").Return(instructors, nil)

	result, err := svc.IsCourseInstructor(ctx, "user-123", "course-123")

	assert.NoError(t, err)
	assert.False(t, result)
	sectionRepo.AssertExpectations(t)
	sectionInstructorRepo.AssertExpectations(t)
}

func TestIsCourseInstructor_NoSections(t *testing.T) {
	svc, _, _, _, sectionRepo, _, _, _ := setupTestService()
	ctx := context.Background()

	sections := []models.Section{}

	sectionRepo.On("GetByCourseID", ctx, "course-123").Return(sections, nil)

	result, err := svc.IsCourseInstructor(ctx, "user-123", "course-123")

	assert.NoError(t, err)
	assert.False(t, result)
	sectionRepo.AssertExpectations(t)
}

// Test IsSectionInstructor
func TestIsSectionInstructor_Success(t *testing.T) {
	svc, _, _, _, _, sectionInstructorRepo, _, _ := setupTestService()
	ctx := context.Background()

	instructors := []models.SectionInstructor{
		{ID: "user-123", Username: "instructor1"},
		{ID: "user-456", Username: "instructor2"},
	}

	sectionInstructorRepo.On("Get", ctx, "section-123").Return(instructors, nil)

	result, err := svc.IsSectionInstructor(ctx, "user-123", "section-123")

	assert.NoError(t, err)
	assert.True(t, result)
	sectionInstructorRepo.AssertExpectations(t)
}

func TestIsSectionInstructor_NotInstructor(t *testing.T) {
	svc, _, _, _, _, sectionInstructorRepo, _, _ := setupTestService()
	ctx := context.Background()

	instructors := []models.SectionInstructor{
		{ID: "user-456", Username: "instructor1"},
	}

	sectionInstructorRepo.On("Get", ctx, "section-123").Return(instructors, nil)

	result, err := svc.IsSectionInstructor(ctx, "user-123", "section-123")

	assert.NoError(t, err)
	assert.False(t, result)
	sectionInstructorRepo.AssertExpectations(t)
}

func TestIsSectionInstructor_EmptyInstructors(t *testing.T) {
	svc, _, _, _, _, sectionInstructorRepo, _, _ := setupTestService()
	ctx := context.Background()

	instructors := []models.SectionInstructor{}

	sectionInstructorRepo.On("Get", ctx, "section-123").Return(instructors, nil)

	result, err := svc.IsSectionInstructor(ctx, "user-123", "section-123")

	assert.NoError(t, err)
	assert.False(t, result)
	sectionInstructorRepo.AssertExpectations(t)
}

func TestIsSectionInstructor_RepositoryError(t *testing.T) {
	svc, _, _, _, _, sectionInstructorRepo, _, _ := setupTestService()
	ctx := context.Background()

	sectionInstructorRepo.On("Get", ctx, "section-123").Return(nil, errors.New("db error"))

	result, err := svc.IsSectionInstructor(ctx, "user-123", "section-123")

	assert.Error(t, err)
	assert.False(t, result)
	sectionInstructorRepo.AssertExpectations(t)
}

// Test IsSectionStudent
func TestIsSectionStudent_Success(t *testing.T) {
	svc, _, _, _, _, _, sectionStudentRepo, _ := setupTestService()
	ctx := context.Background()

	sectionStudent := &models.SectionStudent{
		SectionID: "section-123",
		StudentID: "user-123",
	}

	sectionStudentRepo.On("GetBySectionAndStudentID", ctx, "section-123", "user-123").Return(sectionStudent, nil)

	result, err := svc.IsSectionStudent(ctx, "user-123", "section-123")

	assert.NoError(t, err)
	assert.True(t, result)
	sectionStudentRepo.AssertExpectations(t)
}

func TestIsSectionStudent_NotStudent(t *testing.T) {
	svc, _, _, _, _, _, sectionStudentRepo, _ := setupTestService()
	ctx := context.Background()

	sectionStudentRepo.On("GetBySectionAndStudentID", ctx, "section-123", "user-123").Return(nil, sql.ErrNoRows)

	result, err := svc.IsSectionStudent(ctx, "user-123", "section-123")

	assert.NoError(t, err)
	assert.False(t, result)
	sectionStudentRepo.AssertExpectations(t)
}

func TestIsSectionStudent_RepositoryError(t *testing.T) {
	svc, _, _, _, _, _, sectionStudentRepo, _ := setupTestService()
	ctx := context.Background()

	sectionStudentRepo.On("GetBySectionAndStudentID", ctx, "section-123", "user-123").Return(nil, errors.New("db error"))

	result, err := svc.IsSectionStudent(ctx, "user-123", "section-123")

	assert.Error(t, err)
	assert.False(t, result)
	sectionStudentRepo.AssertExpectations(t)
}

// Test GetSection
func TestGetSection_Success(t *testing.T) {
	svc, _, _, _, sectionRepo, _, _, _ := setupTestService()
	ctx := context.Background()

	raw := &repositories.RawSection{ID: "section-123", Name: "Test Section", CourseID: "course-123"}

	sectionRepo.On("GetByID", ctx, "section-123").Return(raw, nil)

	section, err := svc.GetSection(ctx, "section-123")

	assert.NoError(t, err)
	assert.NotNil(t, section)
	assert.Equal(t, "section-123", section.ID)
	assert.Equal(t, "Test Section", section.Name)
	assert.Equal(t, "course-123", section.CourseID)
	sectionRepo.AssertExpectations(t)
}

func TestGetSection_NotFound(t *testing.T) {
	svc, _, _, _, sectionRepo, _, _, _ := setupTestService()
	ctx := context.Background()

	sectionRepo.On("GetByID", ctx, "section-123").Return(nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusNotFound, Message: "Section not found"}))

	section, err := svc.GetSection(ctx, "section-123")

	assert.Error(t, err)
	assert.Nil(t, section)

	var csErr *cserrors.Error
	assert.True(t, errors.As(err, &csErr))
	assert.Equal(t, http.StatusNotFound, csErr.HttpStatus)
	sectionRepo.AssertExpectations(t)
}

func TestGetSection_RepositoryError(t *testing.T) {
	svc, _, _, _, sectionRepo, _, _, _ := setupTestService()
	ctx := context.Background()

	sectionRepo.On("GetByID", ctx, "section-123").Return(nil, errors.New("db error"))

	section, err := svc.GetSection(ctx, "section-123")

	assert.Error(t, err)
	assert.Nil(t, section)
	sectionRepo.AssertExpectations(t)
}

// Test IsSubmissionOwner
func TestIsSubmissionOwner_Success(t *testing.T) {
	svc, _, _, _, _, _, _, submissionRepo := setupTestService()
	ctx := context.Background()

	submission := &repositories.Submission{
		ID:     "submission-123",
		UserID: "user-123",
	}

	submissionRepo.On("Get", ctx, "submission-123").Return(submission, nil)

	result, err := svc.IsSubmissionOwner(ctx, "user-123", "submission-123")

	assert.NoError(t, err)
	assert.True(t, result)
	submissionRepo.AssertExpectations(t)
}

func TestIsSubmissionOwner_NotOwner(t *testing.T) {
	svc, _, _, _, _, _, _, submissionRepo := setupTestService()
	ctx := context.Background()

	submission := &repositories.Submission{
		ID:     "submission-123",
		UserID: "user-456",
	}

	submissionRepo.On("Get", ctx, "submission-123").Return(submission, nil)

	result, err := svc.IsSubmissionOwner(ctx, "user-123", "submission-123")

	assert.NoError(t, err)
	assert.False(t, result)
	submissionRepo.AssertExpectations(t)
}

func TestIsSubmissionOwner_SubmissionNotFound(t *testing.T) {
	svc, _, _, _, _, _, _, submissionRepo := setupTestService()
	ctx := context.Background()

	submissionRepo.On("Get", ctx, "submission-123").Return(nil, sql.ErrNoRows)

	result, err := svc.IsSubmissionOwner(ctx, "user-123", "submission-123")

	assert.Error(t, err)
	assert.False(t, result)

	var csErr *cserrors.Error
	assert.True(t, errors.As(err, &csErr))
	assert.Equal(t, http.StatusNotFound, csErr.HttpStatus)
	submissionRepo.AssertExpectations(t)
}

func TestIsSubmissionOwner_RepositoryError(t *testing.T) {
	svc, _, _, _, _, _, _, submissionRepo := setupTestService()
	ctx := context.Background()

	submissionRepo.On("Get", ctx, "submission-123").Return(nil, errors.New("db error"))

	result, err := svc.IsSubmissionOwner(ctx, "user-123", "submission-123")

	assert.Error(t, err)
	assert.False(t, result)
	submissionRepo.AssertExpectations(t)
}

// Test GetSubmission
func TestGetSubmission_Success(t *testing.T) {
	svc, _, _, _, _, _, _, submissionRepo := setupTestService()
	ctx := context.Background()

	submission := &repositories.Submission{
		ID:     "submission-123",
		UserID: "user-123",
		Status: models.PASSED,
	}

	submissionRepo.On("Get", ctx, "submission-123").Return(submission, nil)

	result, err := svc.GetSubmission(ctx, "submission-123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "submission-123", result.ID)
	submissionRepo.AssertExpectations(t)
}

func TestGetSubmission_NotFound(t *testing.T) {
	svc, _, _, _, _, _, _, submissionRepo := setupTestService()
	ctx := context.Background()

	submissionRepo.On("Get", ctx, "submission-123").Return(nil, sql.ErrNoRows)

	result, err := svc.GetSubmission(ctx, "submission-123")

	assert.Error(t, err)
	assert.Nil(t, result)
	submissionRepo.AssertExpectations(t)
}

func TestGetSubmission_RepositoryError(t *testing.T) {
	svc, _, _, _, _, _, _, submissionRepo := setupTestService()
	ctx := context.Background()

	submissionRepo.On("Get", ctx, "submission-123").Return(nil, errors.New("db error"))

	result, err := svc.GetSubmission(ctx, "submission-123")

	assert.Error(t, err)
	assert.Nil(t, result)
	submissionRepo.AssertExpectations(t)
}

// Test NewService
func TestNewService(t *testing.T) {
	userRepo := new(mockUserRepo)
	courseRepo := new(mockCourseRepo)
	courseCreatorRepo := new(mockCourseCreatorRepo)
	sectionRepo := new(mockSectionRepo)
	sectionInstructorRepo := new(mockSectionInstructorRepo)
	sectionStudentRepo := new(mockSectionStudentRepo)
	submissionRepo := new(mockSubmissionRepo)

	svc := NewService(
		userRepo,
		courseRepo,
		courseCreatorRepo,
		sectionRepo,
		sectionInstructorRepo,
		sectionStudentRepo,
		submissionRepo,
	)

	assert.NotNil(t, svc)

	// Verify it's the correct type
	_, ok := svc.(*service)
	assert.True(t, ok)
}

// Test multiple roles
func TestIsAdmin_MultipleRoles(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _ := setupTestService()
	ctx := context.Background()

	userData := &repositories.UserData{
		ID:    "user-123",
		Roles: []string{"student", "instructor", "admin"},
	}

	userRepo.On("GetByID", ctx, "user-123").Return(userData, nil)

	result, err := svc.IsAdmin(ctx, "user-123")

	assert.NoError(t, err)
	assert.True(t, result)
	userRepo.AssertExpectations(t)
}

func TestIsInstructor_MultipleRoles(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _ := setupTestService()
	ctx := context.Background()

	userData := &repositories.UserData{
		ID:    "user-123",
		Roles: []string{"student", "instructor"},
	}

	userRepo.On("GetByID", ctx, "user-123").Return(userData, nil)

	result, err := svc.IsInstructor(ctx, "user-123")

	assert.NoError(t, err)
	assert.True(t, result)
	userRepo.AssertExpectations(t)
}

// Test context propagation
func TestService_ContextPropagation(t *testing.T) {
	svc, userRepo, _, _, _, _, _, _ := setupTestService()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userData := &repositories.UserData{
		ID:    "user-123",
		Roles: []string{"admin"},
	}

	userRepo.On("GetByID", ctx, "user-123").Return(userData, nil)

	result, err := svc.IsAdmin(ctx, "user-123")

	assert.NoError(t, err)
	assert.True(t, result)
	userRepo.AssertExpectations(t)
}
