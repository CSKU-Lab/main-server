package services

import (
	"context"
	"net/http"

	"github.com/CSKU-Lab/main-server/configs"
	"github.com/CSKU-Lab/main-server/constants"
	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/converter"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
	"github.com/CSKU-Lab/queue"
	"github.com/google/uuid"
)

// AddStudentsResult reports usernames that were skipped while adding students
// so the caller can surface them to the user.
type AddStudentsResult struct {
	NotFound     []string
	NotStudents  []string
	AlreadyAdded []string
}

type SectionService interface {
	Create(ctx context.Context, req *requests.CreateSection) (string, error)
	UpdateByID(ctx context.Context, ID string, req *requests.UpdateSection, userID string) error
	GetByID(ctx context.Context, ID string) (*models.Section, error)
	GetBySemesterID(ctx context.Context, semesterID string) ([]models.Section, error)
	GetByCourseIDAndSemesterID(ctx context.Context, courseID string, semesterID string) ([]models.Section, error)
	GetRawBySemesterID(ctx context.Context, semesterID string) ([]repositories.RawSection, error)
	DeleteByID(ctx context.Context, ID string, userID string) error
	AddStudents(ctx context.Context, sectionID string, studentUsernames []string) (*AddStudentsResult, error)
	GetStudentsBySectionID(ctx context.Context, sectionID string) ([]models.Student, error)
	RemoveStudents(ctx context.Context, sectionID string, studentIDs []string) error
	SetDefaultLabs(ctx context.Context, sectionID string, courseID string) error
	GetSectionsPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filterParams map[string]string) ([]models.Section, error)
	Count(ctx context.Context, filterParams map[string]string) (int, error)
}

type sectionService struct {
	config                *configs.Config
	uowRepo               repositories.UoWRepository
	repo                  repositories.SectionRepository
	courseRepo            repositories.CourseRepository
	storage               repositories.FileRepository
	sectionInstructorRepo repositories.SectionInstructorRepository
	sectionStudentRepo    repositories.SectionStudentRepository
	userRepo              repositories.User
	semesterRepo          repositories.SemesterRepository
	sectionLogService     SectionLogService
	allowedFilterFields   map[string]bool
	allowedSortFields     map[string]bool
	q                     queue.Queue
}

func NewSectionService(config *configs.Config, repo repositories.SectionRepository, uowRepo repositories.UoWRepository, courseRepo repositories.CourseRepository, sectionInstructorRepo repositories.SectionInstructorRepository, sectionStudentRepo repositories.SectionStudentRepository, storage repositories.FileRepository, userRepo repositories.User, semesterRepo repositories.SemesterRepository, sectionLogService SectionLogService, q queue.Queue) SectionService {
	return &sectionService{
		config:                config,
		repo:                  repo,
		uowRepo:               uowRepo,
		courseRepo:            courseRepo,
		storage:               storage,
		sectionInstructorRepo: sectionInstructorRepo,
		sectionStudentRepo:    sectionStudentRepo,
		userRepo:              userRepo,
		semesterRepo:          semesterRepo,
		sectionLogService:     sectionLogService,
		allowedFilterFields: map[string]bool{
			"student_id":  true,
			"semester_id": true,
			"course_id":   true,
			"name":        true,
			"is_archived": true,
		},
		allowedSortFields: map[string]bool{
			"created_at": true,
			"name":       true,
		},
		q: q,
	}
}

func (s *sectionService) Count(ctx context.Context, filterParams map[string]string) (int, error) {
	filters, err := sanitize.Filters(filterParams, s.allowedFilterFields)
	if err != nil {
		return 0, err
	}

	return s.sectionStudentRepo.Count(ctx, filters)
}

func (s *sectionService) GetSectionsPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filterParams map[string]string) ([]models.Section, error) {
	sanitizedSortBy, err := sanitize.SortBy(sortBy, s.allowedSortFields)
	if err != nil {
		return nil, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid sort by field",
			})
	}

	sanitizedSortOrder, err := sanitize.SortOrder(sortOrder)
	if err != nil {
		return nil, cserrors.New(
			&cserrors.Option{
				HttpStatus: http.StatusBadRequest,
				Message:    "Invalid sort order",
			})
	}

	sanitizedFilters, err := sanitize.Filters(filterParams, s.allowedFilterFields)
	if err != nil {
		return nil, err
	}

	rawSection, err := s.sectionStudentRepo.GetSectionsPagination(ctx, page, limit, sanitizedSortBy, sanitizedSortOrder, sanitizedFilters)
	if err != nil {
		return nil, err
	}

	sections := make([]models.Section, 0, len(rawSection))
	for _, rs := range rawSection {
		semester, err := s.semesterRepo.GetByID(ctx, rs.SemesterID)
		if err != nil {
			return nil, err
		}

		instructors, err := s.sectionInstructorRepo.Get(ctx, rs.ID)
		if err != nil {
			return nil, err
		}

		var banner *string
		if rs.Banner != nil {
			_banner := converter.ToS3Path(s.config, *rs.Banner)
			banner = &_banner
		}

		sections = append(sections, models.Section{
			ID:       rs.ID,
			Name:     rs.Name,
			Banner:   banner,
			CourseID: rs.CourseID,
			Semester: models.SectionSemester{
				ID:   semester.ID,
				Name: semester.Name,
				Type: semester.Type,
			},
			Instructors: instructors,
		})
	}

	return sections, nil
}

func (s *sectionService) SetDefaultLabs(ctx context.Context, sectionID string, courseID string) error {
	err := s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		defaultLabs, err := u.DefaultLab().GetByCourseID(ctx, courseID)
		if err != nil {
			return err
		}

		for _, defaultLab := range defaultLabs {
			labSecID, err := uuid.NewV7()
			if err != nil {
				return cserrors.New(&cserrors.Option{
					HttpStatus: http.StatusInternalServerError,
					Message:    "Cannot generate uuid",
				})
			}

			err = u.LabSection().Create(ctx, repositories.CreateLabSectionParams{
				LabID:     defaultLab.LabID,
				SectionID: sectionID,
				Position:  defaultLab.Position,
				ID:        labSecID.String(),
				Status:    "hidden",
				OpenedAt:  nil,
				ReadonlyAt:  nil,
			})
			if err != nil {
				return err
			}

		}

		err = s.sectionLogService.Create(ctx, sectionID, "Updated default labs for section")
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *sectionService) Create(ctx context.Context, req *requests.CreateSection) (string, error) {
	ID, err := uuid.NewV7()
	if err != nil {
		return "", cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Cannot generate uuid",
		})
	}

	err = s.uowRepo.Execute(ctx, func(suow repositories.UoWInstance) error {
		err := suow.Section().Create(ctx, ID.String(), &repositories.CreateSection{
			Name:       req.Name,
			CourseID:   req.CourseID,
			SemesterID: req.SemesterID,
		})
		if err != nil {
			return err
		}

		for _, instructorID := range req.Instructors {
			err := suow.SectionInstructor().Add(ctx, ID.String(), instructorID)
			if err != nil {
				return err
			}
		}

		if req.Banner != nil {
			image, err := s.storage.UploadFile(ctx, constants.SECTION_BANNER, req.Banner)
			if err != nil {
				return err
			}

			err = suow.Section().UpdateByID(ctx, ID.String(), &repositories.UpdateSection{
				Banner: &image.Path,
			})
			if err != nil {
				s.storage.DeleteFile(ctx, image.Name)
				return err
			}
		}

		if len(req.Students) > 0 {
			studentIds, err := s.userRepo.GetManyByUsername(ctx, req.Students)
			if err != nil {
				return err
			}
			for _, student := range studentIds {
				err := suow.SectionStudent().Add(ctx, ID.String(), student.ID)
				if err != nil {
					return err
				}
			}
		}

		// we can't use service inside uow so we need to use repo directly and generate log here because section didn't exist before
		// we commit the section creation so if log creation fail, section creation will be rolled back
		id, err := uuid.NewV7()
		if err != nil {
			return cserrors.New(&cserrors.Option{
				HttpStatus: http.StatusInternalServerError,
				Message:    "Cannot generate uuid",
			})
		}
		err = suow.SectionLog().Create(ctx, id.String(), ID.String(), "Created section")
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if semester, err := s.semesterRepo.GetByID(ctx, req.SemesterID); err == nil {
		tag := semester.Name
		publishOGEvent(s.q, ogImageEvent{Type: "section", ID: ID.String(), Title: req.Name, Tag: &tag})
	}

	return ID.String(), nil
}

func (s *sectionService) UpdateByID(ctx context.Context, ID string, req *requests.UpdateSection, userID string) error {
	currentSection, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return err
	}
	sectionInstructors, err := s.sectionInstructorRepo.Get(ctx, currentSection.ID)
	if err != nil {
		return err
	}

	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	isAdmin := false
	for _, role := range userData.Roles {
		if role == string(models.ADMIN) {
			isAdmin = true
			break
		}
	}

	isAuthor := isAdmin
	if !isAdmin {
		for _, i := range sectionInstructors {
			if userID == i.ID {
				isAuthor = true
			}
		}
	}
	if !isAuthor {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Message:    "No Permission",
		})
	}

	err = s.uowRepo.Execute(ctx, func(suow repositories.UoWInstance) error {
		if req.Banner != nil {
			imagePath, err := s.storage.UploadFile(ctx, constants.SECTION_BANNER, req.Banner)
			if err != nil {
				return cserrors.New(&cserrors.Option{
					HttpStatus: http.StatusInternalServerError,
					Message:    "Cannot upload image",
				})
			}

			err = suow.Section().UpdateByID(ctx, ID, &repositories.UpdateSection{
				Banner: &imagePath.Path,
			})
			if err != nil {
				return err
			}

			if currentSection.Banner != nil {
				if err := s.storage.DeleteFile(ctx, *currentSection.Banner); err != nil {
					return err
				}
			}
		}

		if len(req.Instructors) > 0 {
			err := suow.SectionInstructor().DeleteBySectionID(ctx, ID)
			if err != nil {
				return err
			}

			for _, instructorID := range req.Instructors {
				err := suow.SectionInstructor().Add(ctx, ID, instructorID)
				if err != nil {
					return err
				}
			}
		}

		err = suow.Section().UpdateByID(ctx, ID, &repositories.UpdateSection{
			Name:       req.Name,
			SemesterID: req.SemesterID,
		})
		if err != nil {
			return err
		}

		return s.sectionLogService.Create(ctx, ID, "Updated section")
	})
	if err != nil {
		return err
	}

	if section, err := s.GetByID(ctx, ID); err == nil {
		tag := section.Semester.Name
		publishOGEvent(s.q, ogImageEvent{Type: "section", ID: ID, Title: section.Name, Tag: &tag})
	}

	return nil
}

func (s *sectionService) GetByID(ctx context.Context, ID string) (*models.Section, error) {
	dbSection, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	section := &models.Section{
		ID:       dbSection.ID,
		Name:     dbSection.Name,
		CourseID: dbSection.CourseID,
	}

	if dbSection.Banner != nil {
		bannerS3Path := converter.ToS3Path(s.config, *dbSection.Banner)
		section.Banner = &bannerS3Path
	}

	instructors, err := s.sectionInstructorRepo.Get(ctx, section.ID)
	if err != nil {
		return nil, err
	}

	semester, err := s.semesterRepo.GetByID(ctx, dbSection.SemesterID)
	if err != nil {
		return nil, err
	}

	section.Semester = models.SectionSemester{
		ID:   semester.ID,
		Name: semester.Name,
		Type: semester.Type,
	}

	section.Instructors = instructors

	return section, nil
}

func (s *sectionService) GetBySemesterID(ctx context.Context, semesterID string) ([]models.Section, error) {
	sections, err := s.repo.GetBySemesterID(ctx, semesterID)
	if err != nil {
		return nil, nil
	}

	for i, section := range sections {
		if sections[i].Banner != nil {
			bannerS3Path := converter.ToS3Path(s.config, *section.Banner)
			sections[i].Banner = &bannerS3Path
		}

		instructors, err := s.sectionInstructorRepo.Get(ctx, section.ID)
		if err != nil {
			return nil, err
		}

		sections[i].Instructors = instructors
	}

	return sections, nil
}

func (s *sectionService) GetByCourseIDAndSemesterID(ctx context.Context, courseID string, semesterID string) ([]models.Section, error) {
	sections, err := s.repo.GetByCourseIDAndSemesterID(ctx, courseID, semesterID)
	if err != nil {
		return nil, nil
	}

	for i, section := range sections {
		if sections[i].Banner != nil {
			bannerS3Path := converter.ToS3Path(s.config, *section.Banner)
			sections[i].Banner = &bannerS3Path
		}

		instructors, err := s.sectionInstructorRepo.Get(ctx, section.ID)
		if err != nil {
			return nil, err
		}

		sections[i].Instructors = instructors
	}

	return sections, nil
}

func (s *sectionService) GetRawBySemesterID(ctx context.Context, semesterID string) ([]repositories.RawSection, error) {
	return s.repo.GetRawBySemesterID(ctx, semesterID)
}

func (s *sectionService) DeleteByID(ctx context.Context, ID string, userID string) error {
	_, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return err
	}

	instructors, err := s.sectionInstructorRepo.Get(ctx, ID)
	if err != nil {
		return err
	}

	userData, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	isAdmin := false
	for _, role := range userData.Roles {
		if role == string(models.ADMIN) {
			isAdmin = true
			break
		}
	}

	isAuthor := isAdmin
	if !isAdmin {
		for _, instructor := range instructors {
			if userID == instructor.ID {
				isAuthor = true
			}
		}
	}
	if !isAuthor {
		return cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusForbidden,
			Message:    "No Permission",
		})
	}
	err = s.repo.DeleteByID(ctx, ID)
	if err != nil {
		return err
	}

	return s.sectionLogService.Create(ctx, ID, "Deleted section")
}

func hasStudentRole(roles []string) bool {
	for _, r := range roles {
		if r == string(models.STUDENT) {
			return true
		}
	}
	return false
}

func (s *sectionService) AddStudents(ctx context.Context, sectionID string, studentUsernames []string) (*AddStudentsResult, error) {
	students, err := s.userRepo.GetManyByUsername(ctx, studentUsernames)
	if err != nil {
		return nil, err
	}

	result := &AddStudentsResult{
		NotFound:     []string{},
		NotStudents:  []string{},
		AlreadyAdded: []string{},
	}

	// Skip usernames that don't exist.
	foundUsernames := make(map[string]bool, len(students))
	for _, student := range students {
		foundUsernames[student.Username] = true
	}
	for _, username := range studentUsernames {
		if !foundUsernames[username] {
			result.NotFound = append(result.NotFound, username)
		}
	}

	// Look up who is already in the section so we can skip duplicates without
	// poisoning the insert transaction with a unique violation.
	existing, err := s.sectionStudentRepo.GetBySectionID(ctx, sectionID)
	if err != nil {
		return nil, err
	}
	existingIDs := make(map[string]bool, len(existing))
	for _, st := range existing {
		existingIDs[st.ID] = true
	}

	// Partition found users: skip non-students and already-added, add the rest.
	toAdd := make([]repositories.UserData, 0, len(students))
	for _, student := range students {
		if !hasStudentRole(student.Roles) {
			result.NotStudents = append(result.NotStudents, student.Username)
			continue
		}
		if existingIDs[student.ID] {
			result.AlreadyAdded = append(result.AlreadyAdded, student.Username)
			continue
		}
		toAdd = append(toAdd, student)
	}

	if len(toAdd) == 0 {
		return result, nil
	}

	err = s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		for _, student := range toAdd {
			if err := u.SectionStudent().Add(ctx, sectionID, student.ID); err != nil {
				return err
			}
		}

		return s.sectionLogService.Create(ctx, sectionID, "Added students to section")
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *sectionService) GetStudentsBySectionID(ctx context.Context, sectionID string) ([]models.Student, error) {
	return s.sectionStudentRepo.GetBySectionID(ctx, sectionID)
}

func (s *sectionService) RemoveStudents(ctx context.Context, sectionID string, studentIDs []string) error {
	return s.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		for _, studentID := range studentIDs {
			err := u.SectionStudent().RemoveBySectionIDAndStudentID(ctx, sectionID, studentID)
			if err != nil {
				return err
			}
		}
		return s.sectionLogService.Create(ctx, sectionID, "Removed students from section")
	})
}
