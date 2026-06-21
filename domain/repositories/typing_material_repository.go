package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type TypingMaterialPayload struct {
	Content    string
	TypingType string
}

type TypingMaterialRepository interface {
	Create(ctx context.Context, materialID string, payload *TypingMaterialPayload) error
	GetByID(ctx context.Context, materialID string) (*models.TypingMaterial, error)
	UpdateByID(ctx context.Context, materialID string, payload *TypingMaterialPayload) error
	DeleteByID(ctx context.Context, materialID string) error
}
