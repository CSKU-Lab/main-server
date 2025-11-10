package repositories

import "context"

type CodeMaterialRepository interface {
	SetDescription(ctx context.Context, materialID string, description string) error
	GetDescription(ctx context.Context, materialID string) (*string, error)
}
