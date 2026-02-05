package repositories

import (
	"context"
)

type UpdateCodeMaterialPayload struct {
	Description   *string
	HideTestCases *bool
}

type CodeMaterialRepository interface {
	Update(ctx context.Context, materialID string, payload *UpdateCodeMaterialPayload) error
	SetTaskID(ctx context.Context, materialID string, taskID string) error
	GetByID(ctx context.Context, materialID string) (*CodeMaterial, error)
}

type CodeMaterial struct {
	ID            string  `db:"material_id"`
	Description   *string `db:"description"`
	TaskID        string  `db:"task_id"`
	HideTestCases bool    `db:"hide_test_cases"`
}
