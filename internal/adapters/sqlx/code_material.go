package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/jmoiron/sqlx"
)

type codeMaterialRepo struct {
	db *sqlx.DB
}

func NewCodeMaterialRepository(db *sqlx.DB) repositories.CodeMaterialRepository {
	return &codeMaterialRepo{db: db}
}

type codeMaterialRecord struct {
	Id            string  `db:"material_id"`
	Description   *string `db:"description"`
	HideTestCases *bool   `db:"hide_test_cases"`
}

func (c *codeMaterialRepo) Update(ctx context.Context, materialID string, payload *repositories.UpdateCodeMaterialPayload) error {
	record := &codeMaterialRecord{
		Id:            materialID,
		Description:   payload.Description,
		HideTestCases: payload.HideTestCases,
	}

	updateFields := getUpdateFields(record)
	if len(updateFields) == 0 {
		return nil
	}

	query := fmt.Sprintf(`UPDATE code_materials SET %s WHERE material_id = :material_id`, updateFields)

	_, err := c.db.NamedExecContext(
		ctx,
		query,
		record,
	)
	return err
}

func (c *codeMaterialRepo) GetByID(ctx context.Context, materialID string) (*repositories.CodeMaterial, error) {
	var codeMat repositories.CodeMaterial
	err := c.db.GetContext(
		ctx,
		&codeMat,
		`SELECT description,task_id,hide_test_cases FROM code_materials WHERE material_id = $1`,
		materialID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, cserrors.New(&cserrors.Option{
				Message:    "code material not found",
				HttpStatus: http.StatusInternalServerError,
			})
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
