package psql

import (
	"github.com/SornchaiTheDev/cs-lab-backend/domain/models"
	"github.com/lib/pq"
)

type Course struct {
	ID       string         `json:"id" db:"id"`
	Name     string         `json:"name" db:"name"`
	Creators pq.StringArray `json:"creators" db:"creators"`
	models.RecordStatus
}

func (c *Course) ToModel() *models.Course {
	return &models.Course{
		ID:   c.ID,
		Name: c.Name,
		Creators: func() []string {
			_creators := make([]string, len(c.Creators))
			for i, creator := range c.Creators {
				_creators[i] = creator
			}
			return _creators
		}(),
		RecordStatus: models.RecordStatus{
			IsDeleted: false,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			DeletedAt: c.DeletedAt,
		},
	}
}
