package repositories

import (
	"context"
	"time"

	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

type MaterialRepository interface {
	Create(ctx context.Context, ID string, createdByUserID string, req *requests.CreateMaterial) error
	GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter) ([]Material, error)
	Count(ctx context.Context, search string, filters []sanitize.Filter) (int, error)
	GetByID(ctx context.Context, ID string) (*Material, error)
	UpdateByID(ctx context.Context, ID string, req *requests.BaseUpdateMaterial) error
	DeleteByID(ctx context.Context, ID string) error
}

type Material struct {
	ID          string
	Name        string
	Type        string
	Visibility  string
	CreatedAt   time.Time
	CreatedBy   string
	AutoScore   int
	ManualScore int
}
