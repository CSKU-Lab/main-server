package services

import (
	"context"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type AffectedEntitiesService interface {
	GetAffectedEntities(ctx context.Context, req *requests.GetAffectedEntities) ([]models.AffectedEntity, error)
}

type affectedEntitiesService struct {
	affectedEntitiesFactory registries.AffectedEntitiesFactory
}

func NewAffectedEntitiesService(affectedEntitiesFactory registries.AffectedEntitiesFactory) AffectedEntitiesService {
	return &affectedEntitiesService{
		affectedEntitiesFactory: affectedEntitiesFactory,
	}
}

func (a *affectedEntitiesService) GetAffectedEntities(ctx context.Context, req *requests.GetAffectedEntities) ([]models.AffectedEntity, error) {
	handler, exists := a.affectedEntitiesFactory.GetHandler(req.Type)
	if !exists {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Message:    "Unsupported entity type",
		})
	}

	res, err := handler.GetByTypeAndID(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
