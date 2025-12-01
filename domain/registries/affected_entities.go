package registries

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type AffectedEntities interface {
	GetByTypeAndID(ctx context.Context, req *requests.GetAffectedEntities) ([]models.AffectedEntity, error)
}

type AffectedEntitiesFactory interface {
	Register(entityType string, handler AffectedEntities)
	GetHandler(entityType string) (AffectedEntities, bool)
}

type affectedEntitiesFactory struct {
	registry map[string]AffectedEntities
}

func NewAffectedEntityFactory() AffectedEntitiesFactory {
	return &affectedEntitiesFactory{
		registry: make(map[string]AffectedEntities),
	}
}

func (a *affectedEntitiesFactory) Register(entityType string, handler AffectedEntities) {
	a.registry[entityType] = handler
}

func (a *affectedEntitiesFactory) GetHandler(entityType string) (AffectedEntities, bool) {
	handler, exists := a.registry[entityType]
	return handler, exists
}
