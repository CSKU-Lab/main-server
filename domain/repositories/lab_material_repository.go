package repositories

import (
	"context"

	"github.com/CSKU-Lab/main-server/internal/requests"
)

type LabMaterialRepository interface {
	Create(ctx context.Context, req *requests.SetLabMaterial) error
}
