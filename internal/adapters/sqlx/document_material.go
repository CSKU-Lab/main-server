package sqlx

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
)

type documentMaterialRepo struct {
	db *sqlx.DB
}

func NewDocumentMaterialRepository(db *sqlx.DB) repositories.DocumentMaterialRepository {
	return &documentMaterialRepo{db: db}
}

func (r *documentMaterialRepo) Create(ctx context.Context, materialID string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO document_materials (material_id) VALUES ($1)`, materialID)
	return err
}

func (r *documentMaterialRepo) GetByID(ctx context.Context, materialID string) (*repositories.DocumentMaterial, error) {
	var doc repositories.DocumentMaterial
	err := r.db.QueryRowxContext(ctx, `SELECT material_id, content FROM document_materials WHERE material_id = $1`, materialID).
		StructScan(&doc)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusNotFound, Message: "document material not found"})
		}
		return nil, err
	}
	return &doc, nil
}

func (r *documentMaterialRepo) UpdateByID(ctx context.Context, materialID string, content string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE document_materials SET content = $2 WHERE material_id = $1`, materialID, content)
	return err
}

func (r *documentMaterialRepo) DeleteByID(ctx context.Context, materialID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM document_materials WHERE material_id = $1`, materialID)
	return err
}
