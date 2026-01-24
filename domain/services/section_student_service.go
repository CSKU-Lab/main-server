package services

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type SectionStudentService interface {
	GetBySectionAndStudentID(ctx context.Context, sectionID string, studentID string) (*models.SectionStudent, error)
}

type sectionStudentService struct {
	sectionStudentRepo repositories.SectionStudentRepository
	sectionRepo        repositories.SectionRepository
	userRepo           repositories.User
}

func NewSectionStudentService(sectionStudentRepo repositories.SectionStudentRepository, sectionRepo repositories.SectionRepository, userRepo repositories.User) SectionStudentService {
	return &sectionStudentService{
		sectionStudentRepo: sectionStudentRepo,
		sectionRepo:        sectionRepo,
		userRepo:           userRepo,
	}
}

func (ss *sectionStudentService) rowExists(ctx context.Context, sectionID string, studentID string) error {
	_, err := ss.sectionRepo.GetByID(ctx, sectionID)
	if err != nil {
		return err
	}
	_, err = ss.userRepo.GetByID(ctx, studentID)
	if err != nil {
		return err
	}
	return nil
}

func (ss *sectionStudentService) GetBySectionAndStudentID(ctx context.Context, sectionID string, studentID string) (*models.SectionStudent, error) {
	err := ss.rowExists(ctx, sectionID, studentID)
	if err != nil {
		return nil, err
	}

	secStudent, err := ss.sectionStudentRepo.GetBySectionAndStudentID(ctx, sectionID, studentID)
	if err != nil {
		return nil, err
	}
	return secStudent, nil
}
