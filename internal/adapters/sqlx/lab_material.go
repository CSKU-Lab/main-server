package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	"github.com/CSKU-Lab/main-server/internal/sanitize"
)

type labMaterialSchema struct {
	ID         string    `db:"id"`
	LabID      string    `db:"lab_id"`
	MaterialID string    `db:"material_id"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

type sqlxLabMaterialRepository struct {
	db instance
}

func NewSqlxLabMaterialRepository(db instance) repositories.LabMaterialRepository {
	return &sqlxLabMaterialRepository{
		db: db,
	}
}

func (lm *sqlxLabMaterialRepository) Create(ctx context.Context, req *requests.SetLabMaterial, id string, labID string) error {
	query := `INSERT INTO lab_materials (lab_id, material_id, id) VALUES ($1, $2, $3)`
	_, err := lm.db.ExecContext(ctx, query, labID, req.MaterialID, id)
	if err != nil {
		return err
	}

	return nil
}

func (lm *sqlxLabMaterialRepository) GetByID(ctx context.Context, labID string, materilaID string) (*models.LabMaterial, error) {
	query := `SELECT id, lab_id, material_id, created_at, updated_at FROM lab_materials WHERE lab_id = $1 AND material_id = $2 AND is_deleted = false`

	labMaterialSchema := &labMaterialSchema{}
	err := lm.db.GetContext(ctx, labMaterialSchema, query, labID, materilaID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusInternalServerError, Message: "LabMaterial not found"})
		}
		return nil, err
	}

	return &models.LabMaterial{
		ID:         labMaterialSchema.ID,
		LabID:      labMaterialSchema.LabID,
		MaterialID: labMaterialSchema.MaterialID,
		CreatedAt:  labMaterialSchema.CreatedAt,
		UpdatedAt:  labMaterialSchema.UpdatedAt,
	}, nil
}

func (lm *sqlxLabMaterialRepository) DeleteByID(ctx context.Context, id string) error {
	query := "UPDATE lab_materials SET is_deleted = true, deleted_at = NOW() WHERE id = $1"
	_, err := lm.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (lm *sqlxLabMaterialRepository) GetByLabID(ctx context.Context, labID string) ([]models.LabMaterial, error) {
	query := `SELECT id, lab_id, material_id, created_at, updated_at FROM lab_materials WHERE lab_id = $1 AND is_deleted = false`

	labMaterialsSchema := []labMaterialSchema{}
	err := lm.db.SelectContext(ctx, &labMaterialsSchema, query, labID)
	if err != nil {
		return nil, err
	}
	labMaterials := make([]models.LabMaterial, 0, len(labMaterialsSchema))
	for _, labMaterial := range labMaterialsSchema {
		labMaterials = append(labMaterials, models.LabMaterial{
			ID:         labMaterial.ID,
			LabID:      labMaterial.LabID,
			MaterialID: labMaterial.MaterialID,
			CreatedAt:  labMaterial.CreatedAt,
			UpdatedAt:  labMaterial.UpdatedAt,
		})
	}
	return labMaterials, nil
}

func (lm *sqlxLabMaterialRepository) GetPagination(ctx context.Context, page int, limit int, sortBy string, sortOrder string, filters []sanitize.Filter) ([]models.LabMaterial, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 1)

	baseQuery := `SELECT id, lab_id, material_id, created_at, updated_at FROM lab_materials WHERE is_deleted = false`
	query := fmt.Sprintf(`%s%s
		ORDER BY %s %s
		OFFSET $%d
		LIMIT $%d`, baseQuery, filterWhereClause, sortBy, sortOrder, len(filterArgs)+1, len(filterArgs)+2)

	args := make([]any, 0, len(filterArgs)+2)
	args = append(args, filterArgs...)
	args = append(args, (page-1)*limit, limit)

	labMaterialsSchema := []labMaterialSchema{}
	err := lm.db.SelectContext(ctx, &labMaterialsSchema, query, args...)
	if err != nil {
		return nil, err
	}
	labMaterials := make([]models.LabMaterial, 0, len(labMaterialsSchema))
	for _, labMaterial := range labMaterialsSchema {
		labMaterials = append(labMaterials, models.LabMaterial{
			ID:         labMaterial.ID,
			LabID:      labMaterial.LabID,
			MaterialID: labMaterial.MaterialID,
			CreatedAt:  labMaterial.CreatedAt,
			UpdatedAt:  labMaterial.UpdatedAt,
		})
	}
	return labMaterials, nil
}

func (lm *sqlxLabMaterialRepository) Count(ctx context.Context, filters []sanitize.Filter) (int, error) {
	filterWhereClause, filterArgs := buildFilterWhereClause(filters, 1)

	baseQuery := `SELECT COUNT(*) FROM lab_materials WHERE is_deleted = false`

	query := baseQuery + filterWhereClause
	var count int
	err := lm.db.GetContext(ctx, &count, query, filterArgs...)
	if err != nil {
		return 0, err
	}

	return count, nil
}
