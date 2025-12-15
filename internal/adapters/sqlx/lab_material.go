package sqlx

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type sqlxLabMaterialRepository struct {
	db instance
}

func NewSqlxLabMaterialRepository(db instance) repositories.LabMaterialRepository {
	return &sqlxLabMaterialRepository{
		db: db,
	}
}

func (lm *sqlxLabMaterialRepository) Create(ctx context.Context, req *requests.SetLabMaterial) error {
	query := `INSERT INTO lab_materials (lab_id, section_id) VALUES ($1, $2)`
	_, err := lm.db.ExecContext(ctx, query, req.LabID, req.MaterialID)
	if err != nil {
		return err
	}

	return nil
}
