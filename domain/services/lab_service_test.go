package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLabRepository is a mock implementation of repositories.LabRepository
type MockLabRepository struct {
	mock.Mock
}

func (m *MockLabRepository) GetByID(ctx context.Context, labID string) (*models.Lab, error) {
	args := m.Called(ctx, labID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Lab), args.Error(1)
}

func (m *MockLabRepository) Create(ctx context.Context, id string, req *requests.CreateLab, userID string) error {
	args := m.Called(ctx, id, req, userID)
	return args.Error(0)
}

func (m *MockLabRepository) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter) ([]models.Lab, error) {
	args := m.Called(ctx, page, limit, search, sortBy, sortOrder, filters)
	return args.Get(0).([]models.Lab), args.Error(1)
}

func (m *MockLabRepository) Count(ctx context.Context, search string, filters []sanitize.Filter) (int, error) {
	args := m.Called(ctx, search, filters)
	return args.Int(0), args.Error(1)
}

func (m *MockLabRepository) UpdateByID(ctx context.Context, labID string, req *requests.BaseUpdateLab) error {
	args := m.Called(ctx, labID, req)
	return args.Error(0)
}

func (m *MockLabRepository) DeleteByID(ctx context.Context, labID string) error {
	args := m.Called(ctx, labID)
	return args.Error(0)
}

func (m *MockLabRepository) ExistsByNameInCourse(ctx context.Context, displayName string, courseID string) (bool, error) {
	args := m.Called(ctx, displayName, courseID)
	return args.Bool(0), args.Error(1)
}

func (m *MockLabRepository) ExistsByNameInCourseExcludingID(ctx context.Context, displayName string, courseID string, excludeLabID string) (bool, error) {
	args := m.Called(ctx, displayName, courseID, excludeLabID)
	return args.Bool(0), args.Error(1)
}

// MockCourseRepository is a mock implementation of repositories.CourseRepository
type MockCourseRepository struct {
	mock.Mock
}

func (m *MockCourseRepository) GetByID(ctx context.Context, courseID string) (*models.Course, error) {
	args := m.Called(ctx, courseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Course), args.Error(1)
}

func (m *MockCourseRepository) Create(ctx context.Context, ID string, c *requests.CreateCourse) error {
	args := m.Called(ctx, ID, c)
	return args.Error(0)
}

func (m *MockCourseRepository) GetPagination(ctx context.Context, page int, pageSize int, search string, sortBy string, sortOrder string, show string) ([]models.Course, error) {
	args := m.Called(ctx, page, pageSize, search, sortBy, sortOrder, show)
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *MockCourseRepository) Count(ctx context.Context, search string, show string) (int, error) {
	args := m.Called(ctx, search, show)
	return args.Int(0), args.Error(1)
}

func (m *MockCourseRepository) UpdateByID(ctx context.Context, ID string, c *requests.UpdateCourse) error {
	args := m.Called(ctx, ID, c)
	return args.Error(0)
}

func (m *MockCourseRepository) DeleteByID(ctx context.Context, ID string) error {
	args := m.Called(ctx, ID)
	return args.Error(0)
}

// MockUserRepository is a mock implementation of repositories.User
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*repositories.UserData, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.UserData), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*repositories.UserData, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.UserData), args.Error(1)
}

func (m *MockUserRepository) GetManyByUsername(ctx context.Context, usernames []string) ([]repositories.UserData, error) {
	args := m.Called(ctx, usernames)
	return args.Get(0).([]repositories.UserData), args.Error(1)
}

func (m *MockUserRepository) GetManyByFindBy(ctx context.Context, data []string, findBy string, role string) ([]repositories.UserData, error) {
	args := m.Called(ctx, data, findBy, role)
	return args.Get(0).([]repositories.UserData), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, ID string) (*repositories.UserData, error) {
	args := m.Called(ctx, ID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.UserData), args.Error(1)
}

func (m *MockUserRepository) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter) ([]repositories.UserData, error) {
	args := m.Called(ctx, page, limit, search, sortBy, sortOrder, filters)
	return args.Get(0).([]repositories.UserData), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context, search string, filters []sanitize.Filter) (int, error) {
	args := m.Called(ctx, search, filters)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user repositories.CreateMultiTypeUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, ID string, user *requests.UpdateUser) error {
	args := m.Called(ctx, ID, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, ID string) error {
	args := m.Called(ctx, ID)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteMany(ctx context.Context, IDs []string) error {
	args := m.Called(ctx, IDs)
	return args.Error(0)
}

// MockUoWInstance is a mock implementation of repositories.UoWInstance
type MockUoWInstance struct {
	mock.Mock
}

func (m *MockUoWInstance) User() repositories.User {
	args := m.Called()
	return args.Get(0).(repositories.User)
}

func (m *MockUoWInstance) UserPassword() repositories.UserPassword {
	args := m.Called()
	return args.Get(0).(repositories.UserPassword)
}

func (m *MockUoWInstance) UserGroup() repositories.UserGroup {
	args := m.Called()
	return args.Get(0).(repositories.UserGroup)
}

func (m *MockUoWInstance) Section() repositories.SectionRepository {
	args := m.Called()
	return args.Get(0).(repositories.SectionRepository)
}

func (m *MockUoWInstance) SectionInstructor() repositories.SectionInstructorRepository {
	args := m.Called()
	return args.Get(0).(repositories.SectionInstructorRepository)
}

func (m *MockUoWInstance) SectionStudent() repositories.SectionStudentRepository {
	args := m.Called()
	return args.Get(0).(repositories.SectionStudentRepository)
}

func (m *MockUoWInstance) SectionLog() repositories.SectionLogRepository {
	args := m.Called()
	return args.Get(0).(repositories.SectionLogRepository)
}

func (m *MockUoWInstance) Course() repositories.CourseRepository {
	args := m.Called()
	return args.Get(0).(repositories.CourseRepository)
}

func (m *MockUoWInstance) CourseCreator() repositories.CourseCreatorRepository {
	args := m.Called()
	return args.Get(0).(repositories.CourseCreatorRepository)
}

func (m *MockUoWInstance) Material() repositories.MaterialRepository {
	args := m.Called()
	return args.Get(0).(repositories.MaterialRepository)
}

func (m *MockUoWInstance) MaterialTag() repositories.WriteMaterialTagRepository {
	args := m.Called()
	return args.Get(0).(repositories.WriteMaterialTagRepository)
}

func (m *MockUoWInstance) Lab() repositories.LabRepository {
	args := m.Called()
	return args.Get(0).(repositories.LabRepository)
}

func (m *MockUoWInstance) LabSection() repositories.LabSectionRepository {
	args := m.Called()
	return args.Get(0).(repositories.LabSectionRepository)
}

func (m *MockUoWInstance) LabMaterial() repositories.LabMaterialRepository {
	args := m.Called()
	return args.Get(0).(repositories.LabMaterialRepository)
}

func (m *MockUoWInstance) DefaultLab() repositories.DefaultLabRepository {
	args := m.Called()
	return args.Get(0).(repositories.DefaultLabRepository)
}

func (m *MockUoWInstance) Submission() repositories.SubmissionRepository {
	args := m.Called()
	return args.Get(0).(repositories.SubmissionRepository)
}

func (m *MockUoWInstance) CodeSubmission() repositories.CodeSubmissionRepository {
	args := m.Called()
	return args.Get(0).(repositories.CodeSubmissionRepository)
}

func (m *MockUoWInstance) CodeSubmissionOutbox() repositories.CodeSubmissionOutboxRepository {
	args := m.Called()
	return args.Get(0).(repositories.CodeSubmissionOutboxRepository)
}

// MockUoWRepository is a mock implementation of repositories.UoWRepository
type MockUoWRepository struct {
	mock.Mock
	MockUoW *MockUoWInstance
}

func (m *MockUoWRepository) Execute(ctx context.Context, cb func(u repositories.UoWInstance) error) error {
	args := m.Called(ctx, cb)
	// If the mock is set up to return nil, actually execute the callback
	if args.Error(0) == nil && cb != nil {
		// Use the configured MockUoWInstance or create a new one
		mockUoW := m.MockUoW
		if mockUoW == nil {
			mockUoW = new(MockUoWInstance)
		}
		return cb(mockUoW)
	}
	return args.Error(0)
}

func TestLabService_Create_Success(t *testing.T) {
	mockLabRepo := new(MockLabRepository)
	mockCourseRepo := new(MockCourseRepository)
	mockUoWInstance := new(MockUoWInstance)
	mockUoWRepo := &MockUoWRepository{MockUoW: mockUoWInstance}

	service := NewLabService(mockLabRepo, mockCourseRepo, mockUoWRepo)

	req := &requests.CreateLab{
		DisplayName: "Test Lab",
		CourseID:    "course-123",
	}
	userID := "user-123"

	// Mock course exists
	mockCourseRepo.On("GetByID", mock.Anything, "course-123").Return(&models.Course{ID: "course-123"}, nil)

	// Mock lab name doesn't exist in course
	mockLabRepo.On("ExistsByNameInCourse", mock.Anything, "Test Lab", "course-123").Return(false, nil)

	// Mock UoW execution - the mock UoWInstance should return the mockLabRepo when Lab() is called
	mockUoWInstance.On("Lab").Return(mockLabRepo)
	mockLabRepo.On("Create", mock.Anything, mock.AnythingOfType("string"), req, userID).Return(nil)
	mockUoWRepo.On("Execute", mock.Anything, mock.Anything).Return(nil)

	labID, err := service.Create(context.Background(), req, userID)

	assert.NoError(t, err)
	assert.NotEmpty(t, labID)
	mockCourseRepo.AssertExpectations(t)
	mockLabRepo.AssertExpectations(t)
	mockUoWRepo.AssertExpectations(t)
	mockUoWInstance.AssertExpectations(t)
}

func TestLabService_Create_DuplicateNameInSameCourse(t *testing.T) {
	mockLabRepo := new(MockLabRepository)
	mockCourseRepo := new(MockCourseRepository)
	mockUoWRepo := new(MockUoWRepository)

	service := NewLabService(mockLabRepo, mockCourseRepo, mockUoWRepo)

	req := &requests.CreateLab{
		DisplayName: "Test Lab",
		CourseID:    "course-123",
	}
	userID := "user-123"

	// Mock course exists
	mockCourseRepo.On("GetByID", mock.Anything, "course-123").Return(&models.Course{ID: "course-123"}, nil)

	// Mock lab name already exists in course
	mockLabRepo.On("ExistsByNameInCourse", mock.Anything, "Test Lab", "course-123").Return(true, nil)

	labID, err := service.Create(context.Background(), req, userID)

	assert.Error(t, err)
	assert.Empty(t, labID)

	var csErr *cserrors.Error
	assert.True(t, errors.As(err, &csErr))
	assert.Equal(t, http.StatusConflict, csErr.HttpStatus)
	assert.Equal(t, cserrors.LabAlreadyExists, csErr.Code)

	mockCourseRepo.AssertExpectations(t)
	mockLabRepo.AssertExpectations(t)
	mockUoWRepo.AssertNotCalled(t, "Execute")
}

func TestLabService_Create_SameNameDifferentCourses(t *testing.T) {
	mockLabRepo := new(MockLabRepository)
	mockCourseRepo := new(MockCourseRepository)
	mockUoWInstance := new(MockUoWInstance)
	mockUoWRepo := &MockUoWRepository{MockUoW: mockUoWInstance}

	service := NewLabService(mockLabRepo, mockCourseRepo, mockUoWRepo)

	req := &requests.CreateLab{
		DisplayName: "Test Lab",
		CourseID:    "course-456",
	}
	userID := "user-123"

	// Mock course exists
	mockCourseRepo.On("GetByID", mock.Anything, "course-456").Return(&models.Course{ID: "course-456"}, nil)

	// Mock lab name doesn't exist in this course (but may exist in other courses)
	mockLabRepo.On("ExistsByNameInCourse", mock.Anything, "Test Lab", "course-456").Return(false, nil)

	// Mock UoW execution
	mockUoWInstance.On("Lab").Return(mockLabRepo)
	mockLabRepo.On("Create", mock.Anything, mock.AnythingOfType("string"), req, userID).Return(nil)
	mockUoWRepo.On("Execute", mock.Anything, mock.Anything).Return(nil)

	labID, err := service.Create(context.Background(), req, userID)

	assert.NoError(t, err)
	assert.NotEmpty(t, labID)
	mockCourseRepo.AssertExpectations(t)
	mockLabRepo.AssertExpectations(t)
	mockUoWRepo.AssertExpectations(t)
	mockUoWInstance.AssertExpectations(t)
}

func TestLabService_UpdateByID_Success(t *testing.T) {
	mockLabRepo := new(MockLabRepository)
	mockCourseRepo := new(MockCourseRepository)
	mockUoWInstance := new(MockUoWInstance)
	mockUoWRepo := &MockUoWRepository{MockUoW: mockUoWInstance}

	service := NewLabService(mockLabRepo, mockCourseRepo, mockUoWRepo)

	labID := "lab-123"
	userID := "user-123"
	req := &requests.BaseUpdateLab{
		DisplayName: "Updated Lab Name",
	}

	// Mock get lab by ID
	mockLabRepo.On("GetByID", mock.Anything, labID).Return(&models.Lab{
		ID:          labID,
		DisplayName: "Old Lab Name",
		CourseID:    "course-123",
	}, nil)

	// Mock check for duplicate name (new name doesn't exist)
	mockLabRepo.On("ExistsByNameInCourseExcludingID", mock.Anything, "Updated Lab Name", "course-123", labID).Return(false, nil)

	// Mock get course
	mockCourseRepo.On("GetByID", mock.Anything, "course-123").Return(&models.Course{
		ID:       "course-123",
		Creators: []models.CourseCreator{{ID: userID}},
	}, nil)

	// Mock UoW for permission check - need to mock User() call
	mockUserRepo := new(MockUserRepository)
	mockUoWInstance.On("User").Return(mockUserRepo)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(&repositories.UserData{
		ID:    userID,
		Roles: []string{"instructor"},
	}, nil)
	mockUoWRepo.On("Execute", mock.Anything, mock.Anything).Return(nil)

	// Mock update
	mockLabRepo.On("UpdateByID", mock.Anything, labID, req).Return(nil)

	err := service.UpdateByID(context.Background(), labID, userID, req)

	assert.NoError(t, err)
	mockLabRepo.AssertExpectations(t)
	mockCourseRepo.AssertExpectations(t)
	mockUoWRepo.AssertExpectations(t)
	mockUoWInstance.AssertExpectations(t)
}

func TestLabService_UpdateByID_DuplicateNameInSameCourse(t *testing.T) {
	mockLabRepo := new(MockLabRepository)
	mockCourseRepo := new(MockCourseRepository)
	mockUoWRepo := new(MockUoWRepository)

	service := NewLabService(mockLabRepo, mockCourseRepo, mockUoWRepo)

	labID := "lab-123"
	userID := "user-123"
	req := &requests.BaseUpdateLab{
		DisplayName: "Existing Lab Name",
	}

	// Mock get lab by ID
	mockLabRepo.On("GetByID", mock.Anything, labID).Return(&models.Lab{
		ID:          labID,
		DisplayName: "Old Lab Name",
		CourseID:    "course-123",
	}, nil)

	// Mock check for duplicate name (new name already exists)
	mockLabRepo.On("ExistsByNameInCourseExcludingID", mock.Anything, "Existing Lab Name", "course-123", labID).Return(true, nil)

	err := service.UpdateByID(context.Background(), labID, userID, req)

	assert.Error(t, err)

	var csErr *cserrors.Error
	assert.True(t, errors.As(err, &csErr))
	assert.Equal(t, http.StatusConflict, csErr.HttpStatus)
	assert.Equal(t, cserrors.LabAlreadyExists, csErr.Code)

	mockLabRepo.AssertExpectations(t)
	mockCourseRepo.AssertNotCalled(t, "GetByID")
	mockUoWRepo.AssertNotCalled(t, "Execute")
	mockLabRepo.AssertNotCalled(t, "UpdateByID")
}

func TestLabService_UpdateByID_SameNameNoCheck(t *testing.T) {
	mockLabRepo := new(MockLabRepository)
	mockCourseRepo := new(MockCourseRepository)
	mockUoWInstance := new(MockUoWInstance)
	mockUoWRepo := &MockUoWRepository{MockUoW: mockUoWInstance}

	service := NewLabService(mockLabRepo, mockCourseRepo, mockUoWRepo)

	labID := "lab-123"
	userID := "user-123"
	req := &requests.BaseUpdateLab{
		DisplayName: "Same Lab Name", // Same as current name
	}

	// Mock get lab by ID
	mockLabRepo.On("GetByID", mock.Anything, labID).Return(&models.Lab{
		ID:          labID,
		DisplayName: "Same Lab Name", // Same name
		CourseID:    "course-123",
	}, nil)

	// No duplicate check needed since name hasn't changed

	// Mock get course
	mockCourseRepo.On("GetByID", mock.Anything, "course-123").Return(&models.Course{
		ID:       "course-123",
		Creators: []models.CourseCreator{{ID: userID}},
	}, nil)

	// Mock UoW for permission check - need to mock User() call
	mockUserRepo := new(MockUserRepository)
	mockUoWInstance.On("User").Return(mockUserRepo)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(&repositories.UserData{
		ID:    userID,
		Roles: []string{"instructor"},
	}, nil)
	mockUoWRepo.On("Execute", mock.Anything, mock.Anything).Return(nil)

	// Mock update
	mockLabRepo.On("UpdateByID", mock.Anything, labID, req).Return(nil)

	err := service.UpdateByID(context.Background(), labID, userID, req)

	assert.NoError(t, err)
	mockLabRepo.AssertExpectations(t)
	mockCourseRepo.AssertExpectations(t)
	mockUoWRepo.AssertExpectations(t)
	mockUoWInstance.AssertExpectations(t)
	// Verify ExistsByNameInCourseExcludingID was NOT called
	mockLabRepo.AssertNotCalled(t, "ExistsByNameInCourseExcludingID")
}

// TestLabService_Create_CourseNotFound tests error handling when course doesn't exist
func TestLabService_Create_CourseNotFound(t *testing.T) {
	mockLabRepo := new(MockLabRepository)
	mockCourseRepo := new(MockCourseRepository)
	mockUoWRepo := new(MockUoWRepository)

	service := NewLabService(mockLabRepo, mockCourseRepo, mockUoWRepo)

	req := &requests.CreateLab{
		DisplayName: "Test Lab",
		CourseID:    "nonexistent-course",
	}
	userID := "user-123"

	// Mock course not found
	mockCourseRepo.On("GetByID", mock.Anything, "nonexistent-course").Return(nil, errors.New("course not found"))

	labID, err := service.Create(context.Background(), req, userID)

	assert.Error(t, err)
	assert.Empty(t, labID)
	assert.Contains(t, err.Error(), "course not found")
	mockCourseRepo.AssertExpectations(t)
	mockLabRepo.AssertNotCalled(t, "ExistsByNameInCourse")
	mockUoWRepo.AssertNotCalled(t, "Execute")
}

// TestLabService_Create_RepositoryError tests error handling when repository fails
func TestLabService_Create_RepositoryError(t *testing.T) {
	mockLabRepo := new(MockLabRepository)
	mockCourseRepo := new(MockCourseRepository)
	mockUoWRepo := new(MockUoWRepository)

	service := NewLabService(mockLabRepo, mockCourseRepo, mockUoWRepo)

	req := &requests.CreateLab{
		DisplayName: "Test Lab",
		CourseID:    "course-123",
	}
	userID := "user-123"

	// Mock course exists
	mockCourseRepo.On("GetByID", mock.Anything, "course-123").Return(&models.Course{ID: "course-123"}, nil)

	// Mock repository error during existence check
	mockLabRepo.On("ExistsByNameInCourse", mock.Anything, "Test Lab", "course-123").Return(false, errors.New("database connection error"))

	labID, err := service.Create(context.Background(), req, userID)

	assert.Error(t, err)
	assert.Empty(t, labID)
	assert.Contains(t, err.Error(), "database connection error")
	mockCourseRepo.AssertExpectations(t)
	mockLabRepo.AssertExpectations(t)
	mockUoWRepo.AssertNotCalled(t, "Execute")
}

// TestLabService_UpdateByID_LabNotFound tests error handling when lab doesn't exist
func TestLabService_UpdateByID_LabNotFound(t *testing.T) {
	mockLabRepo := new(MockLabRepository)
	mockCourseRepo := new(MockCourseRepository)
	mockUoWRepo := new(MockUoWRepository)

	service := NewLabService(mockLabRepo, mockCourseRepo, mockUoWRepo)

	labID := "nonexistent-lab"
	userID := "user-123"
	req := &requests.BaseUpdateLab{
		DisplayName: "Updated Lab Name",
	}

	// Mock lab not found
	mockLabRepo.On("GetByID", mock.Anything, labID).Return(nil, errors.New("lab not found"))

	err := service.UpdateByID(context.Background(), labID, userID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lab not found")
	mockLabRepo.AssertExpectations(t)
	mockCourseRepo.AssertNotCalled(t, "GetByID")
	mockUoWRepo.AssertNotCalled(t, "Execute")
	mockLabRepo.AssertNotCalled(t, "ExistsByNameInCourseExcludingID")
	mockLabRepo.AssertNotCalled(t, "UpdateByID")
}

// TestLabService_UpdateByID_RepositoryError tests error handling during duplicate check
func TestLabService_UpdateByID_RepositoryError(t *testing.T) {
	mockLabRepo := new(MockLabRepository)
	mockCourseRepo := new(MockCourseRepository)
	mockUoWRepo := new(MockUoWRepository)

	service := NewLabService(mockLabRepo, mockCourseRepo, mockUoWRepo)

	labID := "lab-123"
	userID := "user-123"
	req := &requests.BaseUpdateLab{
		DisplayName: "Updated Lab Name",
	}

	// Mock get lab by ID
	mockLabRepo.On("GetByID", mock.Anything, labID).Return(&models.Lab{
		ID:          labID,
		DisplayName: "Old Lab Name",
		CourseID:    "course-123",
	}, nil)

	// Mock repository error during duplicate check
	mockLabRepo.On("ExistsByNameInCourseExcludingID", mock.Anything, "Updated Lab Name", "course-123", labID).Return(false, errors.New("database query failed"))

	err := service.UpdateByID(context.Background(), labID, userID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database query failed")
	mockLabRepo.AssertExpectations(t)
	mockCourseRepo.AssertNotCalled(t, "GetByID")
	mockUoWRepo.AssertNotCalled(t, "Execute")
	mockLabRepo.AssertNotCalled(t, "UpdateByID")
}

// TestLabService_UpdateByID_CourseNotFound tests error handling when course doesn't exist during update
func TestLabService_UpdateByID_CourseNotFound(t *testing.T) {
	mockLabRepo := new(MockLabRepository)
	mockCourseRepo := new(MockCourseRepository)
	mockUoWRepo := new(MockUoWRepository)

	service := NewLabService(mockLabRepo, mockCourseRepo, mockUoWRepo)

	labID := "lab-123"
	userID := "user-123"
	req := &requests.BaseUpdateLab{
		DisplayName: "Updated Lab Name",
	}

	// Mock get lab by ID
	mockLabRepo.On("GetByID", mock.Anything, labID).Return(&models.Lab{
		ID:          labID,
		DisplayName: "Old Lab Name",
		CourseID:    "course-123",
	}, nil)

	// Mock no duplicate name
	mockLabRepo.On("ExistsByNameInCourseExcludingID", mock.Anything, "Updated Lab Name", "course-123", labID).Return(false, nil)

	// Mock course not found
	mockCourseRepo.On("GetByID", mock.Anything, "course-123").Return(nil, errors.New("course not found"))

	err := service.UpdateByID(context.Background(), labID, userID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "course not found")
	mockLabRepo.AssertExpectations(t)
	mockCourseRepo.AssertExpectations(t)
	mockUoWRepo.AssertNotCalled(t, "Execute")
	mockLabRepo.AssertNotCalled(t, "UpdateByID")
}

// TestLabService_UpdateByID_UpdateRepositoryError tests error handling when update fails
func TestLabService_UpdateByID_UpdateRepositoryError(t *testing.T) {
	mockLabRepo := new(MockLabRepository)
	mockCourseRepo := new(MockCourseRepository)
	mockUoWInstance := new(MockUoWInstance)
	mockUoWRepo := &MockUoWRepository{MockUoW: mockUoWInstance}

	service := NewLabService(mockLabRepo, mockCourseRepo, mockUoWRepo)

	labID := "lab-123"
	userID := "user-123"
	req := &requests.BaseUpdateLab{
		DisplayName: "Updated Lab Name",
	}

	// Mock get lab by ID
	mockLabRepo.On("GetByID", mock.Anything, labID).Return(&models.Lab{
		ID:          labID,
		DisplayName: "Old Lab Name",
		CourseID:    "course-123",
	}, nil)

	// Mock no duplicate name
	mockLabRepo.On("ExistsByNameInCourseExcludingID", mock.Anything, "Updated Lab Name", "course-123", labID).Return(false, nil)

	// Mock get course
	mockCourseRepo.On("GetByID", mock.Anything, "course-123").Return(&models.Course{
		ID:       "course-123",
		Creators: []models.CourseCreator{{ID: userID}},
	}, nil)

	// Mock UoW for permission check
	mockUserRepo := new(MockUserRepository)
	mockUoWInstance.On("User").Return(mockUserRepo)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(&repositories.UserData{
		ID:    userID,
		Roles: []string{"instructor"},
	}, nil)
	mockUoWRepo.On("Execute", mock.Anything, mock.Anything).Return(nil)

	// Mock update failure
	mockLabRepo.On("UpdateByID", mock.Anything, labID, req).Return(errors.New("update failed: constraint violation"))

	err := service.UpdateByID(context.Background(), labID, userID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")
	mockLabRepo.AssertExpectations(t)
	mockCourseRepo.AssertExpectations(t)
	mockUoWRepo.AssertExpectations(t)
}
