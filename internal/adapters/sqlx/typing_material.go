package sqlx

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
)

type typingMaterialRepo struct {
	db *sqlx.DB
}

func NewTypingMaterialRepository(db *sqlx.DB) repositories.TypingMaterialRepository {
	return &typingMaterialRepo{db: db}
}

func (r *typingMaterialRepo) Create(ctx context.Context, materialID string, content string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO typing_materials (material_id, content) VALUES ($1, $2)`, materialID, content)
	return err
}

func (r *typingMaterialRepo) GetByID(ctx context.Context, materialID string) (*models.TypingMaterial, error) {
	var content string
	err := r.db.QueryRowxContext(ctx, `SELECT content FROM typing_materials WHERE material_id = $1`, materialID).Scan(&content)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{HttpStatus: http.StatusNotFound, Message: "Typing material not found"})
		}
		return nil, err
	}
	return &models.TypingMaterial{Content: content}, nil
}

func (r *typingMaterialRepo) UpdateByID(ctx context.Context, materialID string, content string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE typing_materials SET content = $2 WHERE material_id = $1`, materialID, content)
	return err
}

func (r *typingMaterialRepo) DeleteByID(ctx context.Context, materialID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM typing_materials WHERE material_id = $1`, materialID)
	return err
}
