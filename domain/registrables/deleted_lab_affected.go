package registrables

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type deletedLabAffected struct {
	labRepo         repositories.LabRepository
	labSectionRepo  repositories.LabSectionRepository
	labMaterialRepo repositories.LabMaterialRepository
	defaultLabRepo  repositories.DefaultLabRepository
	uowRepo         repositories.UoWRepository
}

func NewDeletedLabAffected(
	labRepo repositories.LabRepository,
	labSectionRepo repositories.LabSectionRepository,
	labMaterialRepo repositories.LabMaterialRepository,
	defaultLabRepo repositories.DefaultLabRepository,
	uowRepo repositories.UoWRepository,
) registries.AffectedEntities {
	return &deletedLabAffected{
		labRepo:         labRepo,
		labSectionRepo:  labSectionRepo,
		labMaterialRepo: labMaterialRepo,
		defaultLabRepo:  defaultLabRepo,
		uowRepo:         uowRepo,
	}
}

func (d *deletedLabAffected) GetByTypeAndID(
	ctx context.Context,
	req *requests.GetAffectedEntities,
) ([]models.AffectedEntity, error) {
	_, err := d.labRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	labSecs, err := d.labSectionRepo.GetByLabID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	labMats, err := d.labMaterialRepo.GetByLabID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	defaultLabs, err := d.defaultLabRepo.GetByLabID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	labSecsRes := models.AffectedEntity{
		Type: "Lab Section",
		Data: []models.EntityDetail{},
	}
	labMatsRes := models.AffectedEntity{
		Type: "Lab Material",
		Data: []models.EntityDetail{},
	}
	defaultLabsRes := models.AffectedEntity{
		Type: "Default Lab",
		Data: []models.EntityDetail{},
	}

	d.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		for _, s := range labSecs {
			lab, err := u.Lab().GetByID(ctx, s.LabID)
			if err != nil {
				return err
			}
			sec, err := u.Section().GetByID(ctx, s.SectionID)
			if err != nil {
				return err
			}
			labSecsRes.Data = append(labSecsRes.Data, models.EntityDetail{
				Name:     lab.DisplayName + " - " + sec.Name,
				Children: nil,
			})
		}
		for _, s := range labMats {
			lab, err := u.Lab().GetByID(ctx, s.LabID)
			if err != nil {
				return err
			}
			mat, err := u.Material().GetByID(ctx, s.MaterialID)
			if err != nil {
				return err
			}
			labMatsRes.Data = append(labMatsRes.Data, models.EntityDetail{
				Name:     lab.DisplayName + " - " + mat.Name,
				Children: nil,
			})
		}
		for _, s := range defaultLabs {
			lab, err := u.Lab().GetByID(ctx, s.LabID)
			if err != nil {
				return err
			}
			course, err := u.Course().GetByID(ctx, s.CourseID)
			if err != nil {
				return err
			}
			defaultLabsRes.Data = append(defaultLabsRes.Data, models.EntityDetail{
				Name:     lab.DisplayName + " - " + course.Name,
				Children: nil,
			})
		}

		return nil
	})

	res := []models.AffectedEntity{labSecsRes, labMatsRes, defaultLabsRes}
	return res, nil
}
