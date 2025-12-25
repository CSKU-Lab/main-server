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
}

func NewDeletedLabAffected(
	labRepo repositories.LabRepository,
	labSectionRepo repositories.LabSectionRepository,
	labMaterialRepo repositories.LabMaterialRepository,
	defaultLabRepo repositories.DefaultLabRepository,
) registries.AffectedEntities {
	return &deletedLabAffected{
		labRepo:         labRepo,
		labSectionRepo:  labSectionRepo,
		labMaterialRepo: labMaterialRepo,
		defaultLabRepo:  defaultLabRepo,
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

	for _, s := range labSecs {
		labSecsRes.Data = append(labSecsRes.Data, models.EntityDetail{
			Name:     s.ID,
			Children: nil,
		})
	}
	for _, s := range labMats {
		labMatsRes.Data = append(labMatsRes.Data, models.EntityDetail{
			Name:     s.ID,
			Children: nil,
		})
	}
	for _, s := range defaultLabs {
		defaultLabsRes.Data = append(defaultLabsRes.Data, models.EntityDetail{
			Name:     s.ID,
			Children: nil,
		})
	}

	res := []models.AffectedEntity{labSecsRes, labMatsRes, defaultLabsRes}
	return res, nil
}
