package sqlx

import (
	"context"
	"database/sql"

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
		`INSERT INTO code_materials (material_id, description) VALUES ($1, $2)
		ON CONFLICT (material_id) DO UPDATE SET description = $2`,
		materialID,
		description,
	)
	return err
}

func (c *codeMaterialRepo) GetDescription(ctx context.Context, materialID string) (*string, error) {
	var description string
	err := c.db.GetContext(
		ctx,
		&description,
		`SELECT description FROM code_materials WHERE material_id = $1`,
		materialID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &description, nil
}
