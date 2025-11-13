package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/raw"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

type materialRepository struct {
	db instance
}

func NewMaterialRepository(db instance) repositories.MaterialRepository {
	return &materialRepository{db: db}
}

func (m *materialRepository) Create(ctx context.Context, ID string, createdByUserID string, req *requests.CreateMaterial) error {
	query := `INSERT INTO materials (id, name, type, visibility, created_by) VALUES ($1, $2, $3, $4, $5)`
	_, err := m.db.ExecContext(ctx, query, ID, req.Name, req.Type, req.Visibility, createdByUserID)
	if err != nil {
		return err
	}

	return nil
}

func (m *materialRepository) GetPagination(ctx context.Context, page int, limit int, search string, sortBy string, sortOrder string, filters []sanitize.Filter) ([]raw.Material, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 2)

	baseQuery := `SELECT id, name, type, visibility, created_at, created_by FROM materials
	WHERE (name ILIKE $1)`

	query := fmt.Sprintf(`%s%s
		ORDER BY %s %s
		OFFSET $%d
		LIMIT $%d`, baseQuery, filterWhereClause, sortBy, sortOrder, len(filterArgs)+2, len(filterArgs)+3)

	args := []any{"%" + search + "%"}
	args = append(args, filterArgs...)
	args = append(args, (page-1)*limit, limit)

	materials := []raw.Material{}
	err := m.db.SelectContext(ctx, &materials, query, args...)
	if err != nil {
		return nil, err
	}

	return materials, nil
}

func (m *materialRepository) Count(ctx context.Context, search string, filters []sanitize.Filter) (int, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 2)

	baseQuery := `SELECT COUNT(*) FROM materials
		WHERE (name ILIKE $1)`

	query := fmt.Sprintf(`%s%s`, baseQuery, filterWhereClause)

	args := []any{"%" + search + "%"}
	args = append(args, filterArgs...)

	var count int
	err := m.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m *materialRepository) GetByID(ctx context.Context, ID string) (*raw.Material, error) {
	query := `SELECT id, name, type, visibility, created_at, created_by FROM materials WHERE id = $1`

	mat := &raw.Material{}
	err := m.db.GetContext(ctx, mat, query, ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "Material not found"})
		}
		return nil, err
	}

	return mat, nil
}

func (m *materialRepository) UpdateByID(ctx context.Context, ID string, req *requests.BaseUpdateMaterial) error {
	_, err := m.GetByID(ctx, ID)
	if err != nil {
		return err
	}

	mat := &raw.Material{
		ID:         ID,
		Name:       req.Name,
		Visibility: req.Visibility,
	}

	updateFields := getUpdateFields(mat)
	if len(updateFields) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
	UPDATE materials
	SET %s ,updated_at = NOW()
	WHERE id = :id
	`, updateFields)

	_, err = m.db.NamedExecContext(ctx, query, mat)
	if err != nil {
		return err
	}

	return nil
}

func (m *materialRepository) DeleteByID(ctx context.Context, ID string) error {
	query := `DELETE FROM materials WHERE id = $1`
	_, err := m.db.ExecContext(ctx, query, ID)
	return err
}
