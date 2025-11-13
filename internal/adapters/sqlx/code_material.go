package sqlx

import (
	"context"
	"database/sql"

	"github.com/CSKU-Lab/main-server/domain/raw"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
)

type codeMaterialRepo struct {
	db *sqlx.DB
}

func NewCodeMaterialRepository(db *sqlx.DB) repositories.CodeMaterialRepository {
	return &codeMaterialRepo{db: db}
}

func (c *codeMaterialRepo) SetDescription(ctx context.Context, materialID string, description string) error {
	_, err := c.db.ExecContext(
		ctx,
		`UPDATE code_materials SET description = $1 WHERE material_id = $2`,
		description,
		materialID,
	)
	return err
}

func (c *codeMaterialRepo) GetByID(ctx context.Context, materialID string) (*raw.CodeMaterial, error) {
	var codeMat raw.CodeMaterial
	err := c.db.GetContext(
		ctx,
		&codeMat,
		`SELECT description,task_id FROM code_materials WHERE material_id = $1`,
		materialID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &codeMat, nil
}

func (c *codeMaterialRepo) SetTaskID(ctx context.Context, materialID string, taskID string) error {
	_, err := c.db.ExecContext(
		ctx,
		`INSERT INTO code_materials (material_id, task_id) VALUES ($1, $2)`,
		materialID,
		taskID,
	)
	return err
}
