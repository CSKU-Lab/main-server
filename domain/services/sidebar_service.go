package services

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type SidebarService interface {
	GetSidebar(ctx context.Context, userID string) ([]*models.SidebarSection, error)
}

type sidebarService struct {
	courseRepo        repositories.CourseRepository
	secStudentRepo    repositories.SectionStudentRepository
	labSectionRepo    repositories.LabSectionRepository
	labMaterialRepo   repositories.LabMaterialRepository
	submissionService SubmissionService
}

func NewSidebarService(
	courseRepo repositories.CourseRepository,
	secStudentRepo repositories.SectionStudentRepository,
	labSectionRepo repositories.LabSectionRepository,
	labMaterialRepo repositories.LabMaterialRepository,
	submissionService SubmissionService,
) SidebarService {
	return &sidebarService{
		courseRepo:        courseRepo,
		secStudentRepo:    secStudentRepo,
		labSectionRepo:    labSectionRepo,
		labMaterialRepo:   labMaterialRepo,
		submissionService: submissionService,
	}
}

func (sb *sidebarService) GetSidebar(
	ctx context.Context,
	userID string,
) ([]*models.SidebarSection, error) {
	sections, err := sb.secStudentRepo.GetByStudentID(ctx, userID)
	if err != nil {
		return nil, err
	}

	sidebars := make([]*models.SidebarSection, 0, len(sections))

	for _, section := range sections {
		course, err := sb.courseRepo.GetByID(ctx, section.CourseID)
		if err != nil {
			return nil, err
		}

		sectionItem := &models.SidebarSection{
			ID:         section.ID,
			Name:       section.Name,
			CourseName: course.Name,
			SubItems:   []*models.SidebarLab{},
		}

		labs, err := sb.labSectionRepo.GetBySectionID(ctx, section.ID)
		if err != nil {
			return nil, err
		}

		for _, lab := range labs {
			labItem := &models.SidebarLab{
				ID:       lab.ID,
				Name:     lab.DisplayName,
				SubItems: []*models.SidebarMaterial{},
			}

			mats, err := sb.labMaterialRepo.GetByLabID(ctx, lab.ID)
			if err != nil {
				return nil, err
			}

			for _, mat := range mats {
				labItem.SubItems = append(labItem.SubItems, &models.SidebarMaterial{
					ID:     mat.ID,
					Name:   mat.Name,
					Status: sb.submissionService.GetMaterialStudentStatus(ctx, userID, mat.ID, lab.ID, section.ID),
				})
			}

			sectionItem.SubItems = append(sectionItem.SubItems, labItem)
		}

		sidebars = append(sidebars, sectionItem)
	}

	return sidebars, nil
}
