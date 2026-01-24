package services

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
)

type SidebarService interface {
	GetSidebar(ctx context.Context, userID string) ([]*models.Sidebar, error)
}

type sidebarService struct {
	courseRepo      repositories.CourseRepository
	secStudentRepo  repositories.SectionStudentRepository
	labSectionRepo  repositories.LabSectionRepository
	labMaterialRepo repositories.LabMaterialRepository
}

func NewSidebarService(
	courseRepo repositories.CourseRepository,
	secStudentRepo repositories.SectionStudentRepository,
	labSectionRepo repositories.LabSectionRepository,
	labMaterialRepo repositories.LabMaterialRepository,
) SidebarService {
	return &sidebarService{
		courseRepo:      courseRepo,
		secStudentRepo:  secStudentRepo,
		labSectionRepo:  labSectionRepo,
		labMaterialRepo: labMaterialRepo,
	}
}

func (sb *sidebarService) GetSidebar(
	ctx context.Context,
	userID string,
) ([]*models.Sidebar, error) {
	sections, err := sb.secStudentRepo.GetByStudentID(ctx, userID)
	if err != nil {
		return nil, err
	}

	sidebars := make([]*models.Sidebar, 0, len(sections))

	for _, section := range sections {
		sectionItem := &models.Sidebar{
			ID:       section.ID,
			Name:     section.Name,
			Status:   "NONE",
			SubItems: []*models.Sidebar{},
		}

		labs, err := sb.labSectionRepo.GetBySectionID(ctx, section.ID)
		if err != nil {
			return nil, err
		}

		for _, lab := range labs {
			labItem := &models.Sidebar{
				ID:       lab.ID,
				Name:     lab.DisplayName,
				Status:   "NONE",
				SubItems: []*models.Sidebar{},
			}

			mats, err := sb.labMaterialRepo.GetByLabID(ctx, lab.ID)
			if err != nil {
				return nil, err
			}

			for _, mat := range mats {
				labItem.SubItems = append(labItem.SubItems, &models.Sidebar{
					ID:     mat.ID,
					Name:   mat.Name,
					Status: "NONE",
				})
			}

			sectionItem.SubItems = append(sectionItem.SubItems, labItem)
		}

		sidebars = append(sidebars, sectionItem)
	}

	return sidebars, nil
}
