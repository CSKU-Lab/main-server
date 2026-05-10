package registrables

import (
	"context"
	"errors"
	"net/http"

	"github.com/CSKU-Lab/main-server/domain/cserrors"
	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type DeletedSectionAffected struct {
	sectionStudentRepo repositories.SectionStudentRepository
}

func NewDeletedSectionAffected(sectionStudentRepo repositories.SectionStudentRepository) *DeletedSectionAffected {
	return &DeletedSectionAffected{
		sectionStudentRepo: sectionStudentRepo,
	}
}

func (d *DeletedSectionAffected) GetByTypeAndID(
	ctx context.Context,
	req *requests.GetAffectedEntities,
) ([]models.AffectedEntity, error) {
	if err := validation.Validate(req.ID, is.UUID); err != nil {
		return nil, cserrors.New(&cserrors.Option{
			HttpStatus: http.StatusBadRequest,
			Code:       errors.New("INVALID_UUID"),
			Message:    "The provided section_id is not a valid UUID",
		})
	}

	res := []models.AffectedEntity{}

	students, err := d.sectionStudentRepo.GetBySectionID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	res = append(res, models.AffectedEntity{
		Type: "Students",
		Data: []models.EntityDetail{},
	})

	for _, student := range students {
		res[0].Data = append(res[0].Data, models.EntityDetail{
			Name:     student.DisplayName,
			Children: nil,
		})
	}

	return res, nil
}
