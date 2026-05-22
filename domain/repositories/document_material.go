package repositories

import "context"

type DocumentMaterialRepository interface {
	Create(ctx context.Context, materialID string) error
	GetByID(ctx context.Context, materialID string) (*DocumentMaterial, error)
	UpdateByID(ctx context.Context, materialID string, content string) error
	DeleteByID(ctx context.Context, materialID string) error
}

type DocumentMaterial struct {
	MaterialID string  `db:"material_id"`
	Content    *string `db:"content"`
}
