package registrables

import (
	"context"
	"strings"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type DeletedLabSectionAffected struct {
	labMaterialRepo repositories.LabMaterialRepository
	uowRepo         repositories.UoWRepository
}

func NewDeletedLabSectionAffected(
	labMaterialRepo repositories.LabMaterialRepository,
	uowRepo repositories.UoWRepository,
) *DeletedLabSectionAffected {
	return &DeletedLabSectionAffected{
		labMaterialRepo: labMaterialRepo,
		uowRepo:         uowRepo,
	}
}

// req.ID is formatted as "labID:sectionID"
func (d *DeletedLabSectionAffected) GetByTypeAndID(
	ctx context.Context,
	req *requests.GetAffectedEntities,
) ([]models.AffectedEntity, error) {
	parts := strings.SplitN(req.ID, ":", 2)
	labID := parts[0]
	sectionID := ""
	if len(parts) == 2 {
		sectionID = parts[1]
	}

	labMats, err := d.labMaterialRepo.GetByLabID(ctx, labID)
	if err != nil {
		return nil, err
	}

	labMatsRes := models.AffectedEntity{
		Type: "Lab Materials",
		Data: []models.EntityDetail{},
	}
	for _, m := range labMats {
		name := m.MaterialID
		if m.MaterialData != nil {
			name = m.MaterialData.Name
		}
		labMatsRes.Data = append(labMatsRes.Data, models.EntityDetail{
			Name:     name,
			Children: nil,
		})
	}

	studentsRes := models.AffectedEntity{
		Type:    "Students",
		Data:    []models.EntityDetail{},
	}

	if sectionID != "" {
		d.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
			students, err := u.SectionStudent().GetBySectionID(ctx, sectionID)
			if err != nil {
				return err
			}
			for _, s := range students {
				studentsRes.Data = append(studentsRes.Data, models.EntityDetail{
					Name:     s.DisplayName,
					Children: nil,
				})
			}
			return nil
		})
	}

	return []models.AffectedEntity{labMatsRes, studentsRes}, nil
}
