package registrables

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/registries"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type deletedCourseAffected struct {
	courseRepo  repositories.CourseRepository
	sectionRepo repositories.SectionRepository
}

func NewDeletedCourseAffected(courseRepo repositories.CourseRepository, sectionRepo repositories.SectionRepository) registries.AffectedEntities {
	return &deletedCourseAffected{
		courseRepo:  courseRepo,
		sectionRepo: sectionRepo,
	}
}

func (d *deletedCourseAffected) GetByTypeAndID(
	ctx context.Context,
	req *requests.GetAffectedEntities,
) ([]models.AffectedEntity, error) {
	_, err := d.courseRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	sections, err := d.sectionRepo.GetByCourseID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	sectionRes := models.AffectedEntity{
		Type: "Section",
		Data: []models.EntityDetail{},
	}

	for _, s := range sections {
		sectionRes.Data = append(sectionRes.Data, models.EntityDetail{
			Name:     s.Name,
			Children: nil,
		})
	}

	res := []models.AffectedEntity{sectionRes}
	return res, nil
}
