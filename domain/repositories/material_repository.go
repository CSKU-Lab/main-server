package repositories

import (
	"context"
	"time"

	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

// VisibilityFilter controls which materials a viewer can see.
// nil = no filter (admin). OnlyPublic = public only (student).
// ViewerID set = public + own (instructor).
type VisibilityFilter struct {
	ViewerID   string
	OnlyPublic bool
}

type MaterialRepository interface {
	Create(ctx context.Context, ID string, courseID string, createdByUserID string, forkedFromMaterialID *string, req *requests.CreateMaterial) error
	GetPagination(ctx context.Context, courseID string, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter, visibility *VisibilityFilter) ([]Material, error)
	Count(ctx context.Context, courseID string, search string, filters []sanitize.Filter, visibility *VisibilityFilter) (int, error)
	GetByID(ctx context.Context, ID string) (*Material, error)
	UpdateByID(ctx context.Context, ID string, req *requests.BaseUpdateMaterial) error
	DeleteByID(ctx context.Context, ID string) error
}

type Material struct {
	ID                   string
	CourseID             string
	ForkedFromMaterialID *string
	Name                 string
	Type                 string
	Visibility           string
	CreatedAt            time.Time
	CreatedBy            string
	AutoScore            int
	ManualScore          int
}
