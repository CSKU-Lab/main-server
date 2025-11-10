package registries

import (
	"context"

	"github.com/CSKU-Lab/main-server/internal/requests"
)

type MaterialRegisterable interface {
	Execute(ctx context.Context, ID string, req *requests.UpdateMaterial) error
	Response(ctx context.Context, ID string) (any, error)
}

type Material interface {
	Register(materialType string, handler MaterialRegisterable)
	GetHandler(materialType string) (MaterialRegisterable, bool)
}

type materialRegistry struct {
	registry map[string]MaterialRegisterable
}

func NewMaterialRegistry() Material {
	return &materialRegistry{
		registry: make(map[string]MaterialRegisterable),
	}
}

func (mr *materialRegistry) Register(materialType string, handler MaterialRegisterable) {
	mr.registry[materialType] = handler
}

func (mr *materialRegistry) GetHandler(materialType string) (MaterialRegisterable, bool) {
	handler, exists := mr.registry[materialType]
	return handler, exists
}
