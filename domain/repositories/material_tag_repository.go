package repositories

import "context"

type WriteMaterialTagRepository interface {
	SetTags(ctx context.Context, materialID string, tags []string) error
}

type ReadMaterialTagRepository interface {
	GetTags(ctx context.Context, materialID string) ([]string, error)
}
