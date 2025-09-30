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
	"github.com/google/uuid"
)

type SectionService interface {
	Create(ctx context.Context, req *requests.CreateSection) (string, error)
	UpdateByID(ctx context.Context, ID string, req *requests.UpdateSection) error
	GetByID(ctx context.Context, ID string) (*models.Section, error)
	GetBySemesterID(ctx context.Context, semesterID string) ([]models.Section, error)
	DeleteByID(ctx context.Context, ID string) error
}

type sectionService struct {
	config                *configs.Config
	uowRepo               repositories.SectionUoWRepository
	repo                  repositories.SectionRepository
	courseRepo            repositories.CourseRepository
	storage               repositories.FileRepository
	sectionInstructorRepo repositories.SectionInstructorRepository
	sectionStudentRepo    repositories.SectionStudentRepository
	userRepo              repositories.User
}

func NewSectionService(config *configs.Config, repo repositories.SectionRepository, uowRepo repositories.SectionUoWRepository, courseRepo repositories.CourseRepository, sectionInstructorRepo repositories.SectionInstructorRepository, sectionStudentRepo repositories.SectionStudentRepository, storage repositories.FileRepository, userRepo repositories.User) SectionService {
	return &sectionService{
		config:                config,
		repo:                  repo,
		uowRepo:               uowRepo,
		courseRepo:            courseRepo,
		storage:               storage,
		sectionInstructorRepo: sectionInstructorRepo,
		sectionStudentRepo:    sectionStudentRepo,
		userRepo:              userRepo,
	}
}

func (s *sectionService) Create(ctx context.Context, req *requests.CreateSection) (string, error) {
	ID, err := uuid.NewV7()
	if err != nil {
		return "", cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Cannot generate uuid",
		})
	}

	err = s.uowRepo.Execute(ctx, func(suow repositories.SectionUoWInstance) error {
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

		return nil
	})

	return ID.String(), err

}

func (s *sectionService) UpdateByID(ctx context.Context, ID string, req *requests.UpdateSection) error {
	currentSection, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return err
	}

	return s.uowRepo.Execute(ctx, func(suow repositories.SectionUoWInstance) error {
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

		if len(req.Students) > 0 {
			err := suow.SectionStudent().DeleteBySectionID(ctx, ID)
			if err != nil {
				return err
			}

			for _, studentID := range req.Students {
				err := suow.SectionStudent().Add(ctx, ID, studentID)
				if err != nil {
					return err
				}
			}
		}

		return suow.Section().UpdateByID(ctx, ID, &repositories.UpdateSection{
			Name:       req.Name,
			SemesterID: req.SemesterID,
		})
	})
}

func (s *sectionService) GetByID(ctx context.Context, ID string) (*models.Section, error) {
	section, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	if section.Banner != nil {
		bannerS3Path := converter.ToS3Path(s.config, *section.Banner)
		section.Banner = &bannerS3Path
	}

	instructors, err := s.sectionInstructorRepo.Get(ctx, section.ID)
	if err != nil {
		return nil, err
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
func (s *sectionService) DeleteByID(ctx context.Context, ID string) error {
	return s.repo.DeleteByID(ctx, ID)
}
