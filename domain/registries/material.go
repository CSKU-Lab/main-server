package registries

import (
	"context"

	"github.com/CSKU-Lab/main-server/internal/requests"
)

type MaterialScores struct {
	AutoScore   int
	ManualScore int
}

type MaterialRegisterable interface {
	Create(ctx context.Context, matID string, req *requests.CreateMaterial, rawReq []byte) error
	GetByID(ctx context.Context, ID string) (any, error)
	CalculateScores(rawReq []byte) (*MaterialScores, error)
	UpdateByID(ctx context.Context, ID string, req *requests.BaseUpdateMaterial, rawReq []byte) error
	DeleteByID(ctx context.Context, ID string) error
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
