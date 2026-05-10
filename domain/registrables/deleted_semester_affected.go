package registrables

import (
	"context"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/CSKU-Lab/main-server/domain/repositories"
	"github.com/CSKU-Lab/main-server/internal/requests"
)

type DeletedSemesterAffected struct {
	semesterRepo repositories.SemesterRepository
	sectionRepo  repositories.SectionRepository
	courseRepo   repositories.CourseRepository
}

func NewDeletedSemesterAffected(semesterRepo repositories.SemesterRepository, sectionRepo repositories.SectionRepository, courseRepo repositories.CourseRepository) *DeletedSemesterAffected {
	return &DeletedSemesterAffected{
		semesterRepo: semesterRepo,
		sectionRepo:  sectionRepo,
		courseRepo:   courseRepo,
	}
}

func (d *DeletedSemesterAffected) GetByTypeAndID(
	ctx context.Context,
	req *requests.GetAffectedEntities,
) ([]models.AffectedEntity, error) {
	_, err := d.semesterRepo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	sections, err := d.sectionRepo.GetRawBySemesterID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	courseWithSectionsMap := make(map[string][]models.EntityDetail)

	for _, section := range sections {
		courseWithSectionsMap[section.CourseID] = append(
			courseWithSectionsMap[section.CourseID],
			models.EntityDetail{
				Name:     section.Name,
				Children: nil,
			},
		)
	}

	courseRes := models.AffectedEntity{
		Type: "Course",
		Data: []models.EntityDetail{},
	}

	for courseID, sectionEntity := range courseWithSectionsMap {
		course, err := d.courseRepo.GetByID(ctx, courseID)
		if err != nil {
			return nil, err
		}

		courseRes.Data = append(courseRes.Data, models.EntityDetail{
			Name: course.Name,
			Children: []models.AffectedEntity{{
				Type: "Section",
				Data: sectionEntity,
			}},
		})
	}

	res := []models.AffectedEntity{courseRes}
	return res, nil
}
