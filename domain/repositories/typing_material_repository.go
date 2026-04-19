package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
)

type TypingMaterialRepository interface {
	Create(ctx context.Context, materialID string, content string) error
	GetByID(ctx context.Context, materialID string) (*models.TypingMaterial, error)
	UpdateByID(ctx context.Context, materialID string, content string) error
	DeleteByID(ctx context.Context, materialID string) error
}
