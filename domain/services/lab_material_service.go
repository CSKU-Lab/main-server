package services

import (
	"context"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type LabMaterialService interface {
	Create(ctx context.Context, req *requests.SetLabMaterial, userID string) error
}

type labMaterialService struct {
	labMaterialRepo repositories.LabMaterialRepository
	uowRepo         repositories.UoWRepository
}

func NewLabMaterialService(labMaterialRepo repositories.LabMaterialRepository, uowRepo repositories.UoWRepository) LabMaterialService {
	return &labMaterialService{
		labMaterialRepo: labMaterialRepo,
		uowRepo:         uowRepo,
	}
}

func (lm *labMaterialService) mutationPermission(ctx context.Context, userID string, labID string) error {
	err := lm.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		user, err := u.User().GetByID(ctx, userID)
		if err != nil {
			return err
		}

		for _, role := range user.Roles {
			if role == string(models.ADMIN) {
				return nil
			}
		}

		lab, err := u.Lab().GetByID(ctx, labID)
		if err != nil {
			return err
		}

		courseCreator, err := u.CourseCreator().GetCreators(ctx, lab.CourseID)
		if err != nil {
			return err
		}

		for _, creator := range courseCreator {
			if creator.ID == userID {
				return nil
			}
		}

		return cserrors.New(&cserrors.Option{
			Message:    "No Permission",
			HttpStatus: http.StatusForbidden,
		})
	})
	if err != nil {
		return err
	}
	return nil
}

func (lm *labMaterialService) rowExists(ctx context.Context, labID string, materialID string) error {
	err := lm.uowRepo.Execute(ctx, func(u repositories.UoWInstance) error {
		_, err := u.Lab().GetByID(ctx, labID)
		if err != nil {
			return err
		}

		_, err = u.Material().GetByID(ctx, materialID)
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

func (lm *labMaterialService) Create(ctx context.Context, req *requests.SetLabMaterial, userID string) error {
	err := lm.rowExists(ctx, req.LabID, req.MaterialID)
	if err != nil {
		return err
	}

	err = lm.mutationPermission(ctx, userID, req.LabID)
	if err != nil {
		return err
	}
	return lm.labMaterialRepo.Create(ctx, req)
}
