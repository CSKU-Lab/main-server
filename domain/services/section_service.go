package services

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/SornchaiTheDev/cs-lab-backend/constants"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/cserrors"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/SornchaiTheDev/cs-lab-backend/domain/repositories"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/requests"
	"github.com/SornchaiTheDev/cs-lab-backend/internal/transaction"
	"github.com/google/uuid"
)

type SectionService interface {
	Create(ctx context.Context, req *requests.CreateSection) (*models.Section, error)
	UpdateByID(ctx context.Context, ID string, req *requests.Section) (*models.Section, error)
	GetByID(ctx context.Context, ID string) (*models.Section, error)
	GetBySemesterID(ctx context.Context, semesterID string) ([]models.Section, error)
	DeleteByID(ctx context.Context, ID string) error
}

type sectionService struct {
	repo       repositories.SectionRepository
	courseRepo repositories.CourseRepository
	storage    repositories.FileRepository
}

func NewSectionService(repo repositories.SectionRepository, courseRepo repositories.CourseRepository, storage repositories.FileRepository) SectionService {
	return &sectionService{
		repo:       repo,
		courseRepo: courseRepo,
		storage:    storage,
	}
}

func (s *sectionService) Create(ctx context.Context, req *requests.CreateSection) (*models.Section, error) {
	ID, err := uuid.NewV7()
	if err != nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusInternalServerError,
			Message:    "Cannot generate uuid",
		})
	}

	section := &models.Section{
		ID:        ID.String(),
		Name:      req.Name,
		Image:     nil,
		StartedAt: req.StartedAt,
		EndedAt:   req.EndedAt,
	}

	tr := transaction.New(&transaction.Option{
		RetryCount: 3,
		RetryDelay: 1 * time.Second,
	})

	var image *models.UploadedFile
	err = tr.Execute(
		tr.Step().CommitWith(func() error {
			_, err := s.courseRepo.GetByID(ctx, req.CourseID)
			if err != nil {
				var csErr *cserrors.Error
				if errors.As(err, &csErr) && csErr.HttpStatus == http.StatusInternalServerError {
					return cserrors.New(&cserrors.Option{
						HttpStatus: http.StatusInternalServerError,
						Message:    "Cannot find course",
					})
				}
			}
			return s.repo.Create(ctx, section, req.CourseID, req.SemesterID)
		}).RollbackWith(func() error {
			return s.repo.DeleteByID(ctx, ID.String())
		}),
		tr.Step().CommitWith(func() error {
			if req.Image == nil {
				return nil
			}

			image, err = s.storage.UploadFile(ctx, constants.SECTION_BANNER, req.Image)
			return err
		}).RollbackWith(func() error {
			return s.storage.DeleteFile(ctx, image.Name)
		}),
		tr.Step().CommitWith(func() error {
			if image == nil {
				return nil
			}

			return s.repo.UpdateByID(ctx, &models.Section{
				ID:    ID.String(),
				Image: &image.Path,
			})
		}),
	)
	if err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, ID.String())
}

func (s *sectionService) UpdateByID(ctx context.Context, ID string, req *requests.Section) (*models.Section, error) {
	tr := transaction.New(&transaction.Option{
		RetryCount: 3,
		RetryDelay: 1 * time.Second,
	})

	currentSection, err := s.repo.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	updatedSection := &models.Section{
		ID:        ID,
		Name:      req.Name,
		StartedAt: req.StartedAt,
		EndedAt:   req.EndedAt,
	}

	var image *models.UploadedFile
	tr.Execute(
		tr.Step().CommitWith(func() error {
			if req.Image != nil {
				image, err := s.storage.UploadFile(ctx, constants.SECTION_BANNER, req.Image)
				if err != nil {
					return cserrors.New(&cserrors.Option{
						HttpStatus: http.StatusInternalServerError,
						Message:    "Cannot upload image",
					})
				}
				updatedSection.Image = &image.Path
			}
			return nil
		}).RollbackWith(func() error {
			if image != nil {
				if err := s.storage.DeleteFile(ctx, image.Name); err != nil {
					return err
				}
			}
			return nil
		}),
		tr.Step().CommitWith(func() error {
			return s.repo.UpdateByID(ctx, updatedSection)
		}).RollbackWith(func() error {
			return s.repo.UpdateByID(ctx, currentSection)
		}),
		tr.Step().CommitWith(func() error {
			if currentSection.Image != nil {
				return s.storage.DeleteFile(ctx, *currentSection.Image)
			}
			return nil
		}),
	)

	return s.repo.GetByID(ctx, ID)
}

func (s *sectionService) GetByID(ctx context.Context, ID string) (*models.Section, error) {
	return s.repo.GetByID(ctx, ID)
}
func (s *sectionService) GetBySemesterID(ctx context.Context, semesterID string) ([]models.Section, error) {
	return s.repo.GetBySemesterID(ctx, semesterID)
}
func (s *sectionService) DeleteByID(ctx context.Context, ID string) error {
	return s.repo.DeleteByID(ctx, ID)
}
