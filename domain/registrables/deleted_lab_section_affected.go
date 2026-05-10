package registrables

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type DeletedLabSectionAffected struct{}

func NewDeletedLabSectionAffected() *DeletedLabSectionAffected {
	return &DeletedLabSectionAffected{}
}

func (d *DeletedLabSectionAffected) GetByTypeAndID(
	ctx context.Context,
	req *requests.GetAffectedEntities,
) ([]models.AffectedEntity, error) {
	mockData := []models.AffectedEntity{
		{
			Type: "Lab Materials",
			Data: []models.EntityDetail{
				{Name: "Lab Material 1 (Mock)", Children: nil},
				{Name: "Lab Material 2 (Mock)", Children: nil},
			},
		},
		{
			Type: "Students",
			Data: []models.EntityDetail{
				{Name: "Student A (Mock)", Children: nil},
				{Name: "Student B (Mock)", Children: nil},
				{Name: "Student C (Mock)", Children: nil},
			},
		},
		{
			Type: "Lab Records",
			Data: []models.EntityDetail{
				{Name: "Lab Record 1 (Mock)", Children: nil},
			},
		},
	}

	return mockData, nil
}
