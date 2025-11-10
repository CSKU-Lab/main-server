package registrables

import (
	"context"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type codeMaterial struct {
	repo repositories.CodeMaterialRepository
}

type CodeMaterialPayload struct {
	Description *string `json:"description"`
}

func (c *CodeMaterialPayload) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Description, validation.Required),
	)
}

func NewCodeMaterial(repo repositories.CodeMaterialRepository) registries.MaterialRegisterable {
	return &codeMaterial{repo: repo}
}

func parsePayload(payload any) (*CodeMaterialPayload, error) {
	payloadMap, ok := payload.(map[string]any)
	if !ok {
		return nil, cserrors.New(&cserrors.Option{
			Message:    "invalid payload format",
			HttpStatus: http.StatusBadRequest,
		})
	}

	description, ok := payloadMap["description"].(string)
	if !ok {
		return nil, cserrors.New(&cserrors.Option{
			Message:    "invalid description field in payload",
			HttpStatus: http.StatusBadRequest,
		})
	}

	return &CodeMaterialPayload{
		Description: &description,
	}, nil
}

func (c *codeMaterial) Execute(ctx context.Context, ID string, req *requests.UpdateMaterial) error {
	payload, err := parsePayload(req.Payload)
	if err != nil {
		return err
	}

	err = c.repo.SetDescription(ctx, ID, *payload.Description)
	if err != nil {
		return err
	}

	return nil
}

func (c *codeMaterial) Response(ctx context.Context, ID string) (any, error) {
	description, err := c.repo.GetDescription(ctx, ID)
	if err != nil {
		return nil, err
	}

	res := &CodeMaterialPayload{
		Description: description,
	}

	return res, nil
}
