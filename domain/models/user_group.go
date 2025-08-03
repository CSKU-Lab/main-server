package models

import "github.com/SornchaiTheDev/cs-lab-backend/internal/responses"

type UserGroup struct {
	ID   string
	Name string
}

func (u *UserGroup) ToResponse() *responses.UserGroup {
	return &responses.UserGroup{
		ID:   u.ID,
		Name: u.Name,
	}
}
