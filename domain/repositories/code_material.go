package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/raw"
)

type CodeMaterialRepository interface {
	SetDescription(ctx context.Context, materialID string, description string) error
	SetTaskID(ctx context.Context, materialID string, taskID string) error
	GetByID(ctx context.Context, materialID string) (*raw.CodeMaterial, error)
}
