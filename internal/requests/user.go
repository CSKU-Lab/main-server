package requests

import (
	"github.com/CSKU-Lab/main-server/domain/models"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type GetInvalidUsers struct {
	Role   string   `json:"role"`
	FindBy string   `json:"find_by"`
	Users  []string `json:"users"`
}

func (c *GetInvalidUsers) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Role, validation.Required, validation.In("admin", "instructor", "student").Error("must be one of 'admin', 'instructor', 'student'")),
		validation.Field(&c.FindBy, validation.Required),
		validation.Field(&c.Users, validation.Required, validation.Length(1, 0)),
	)
}

type CreateMultiTypeUser struct {
	Username    string          `json:"username"`
	DisplayName string          `json:"display_name"`
	Roles       []string        `json:"roles"`
	Type        models.UserType `json:"type"`
	Email       *string         `json:"email"`
	Password    *string         `json:"password"`
	GroupID     *string         `json:"group_id"`
	Group       *string         `json:"group"`
}

func (c *CreateMultiTypeUser) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Username, validation.Required),
		validation.Field(&c.DisplayName, validation.Required),
		validation.Field(&c.Roles, validation.Required, validation.Each(validation.In("admin", "instructor", "student").Error("must be one of 'admin', 'instructor', 'student'"))),
		validation.Field(&c.Type),
		validation.Field(&c.Password, validation.When(c.Type == "credential", validation.Required, validation.Length(8, 0)).Else(validation.Nil)),
		validation.Field(&c.Email, validation.When(c.Type == "oauth", validation.Required.Error("required for oauth user")), validation.When(c.Email != nil, is.Email)),
		validation.Field(&c.GroupID, validation.NilOrNotEmpty, is.UUID),
		validation.Field(&c.Group, validation.NilOrNotEmpty),
	)
}

type UpdateUser struct {
	Username     string   `json:"username"`
	DisplayName  string   `json:"display_name"`
	Roles        []string `json:"roles"`
	Email        *string  `json:"email"`
	Password     *string  `json:"password"`
	ProfileImage *string  `json:"profile_image"`
	GroupID      *string  `json:"group_id"`
}

func (u *UpdateUser) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.Username, validation.NilOrNotEmpty),
		validation.Field(&u.DisplayName, validation.NilOrNotEmpty),
		validation.Field(&u.Roles, validation.NilOrNotEmpty, validation.Each(validation.In("admin", "instructor", "student"))),
		validation.Field(&u.Email, validation.NilOrNotEmpty, is.Email),
		validation.Field(&u.Password, validation.NilOrNotEmpty, validation.Length(8, 0)),
		validation.Field(&u.ProfileImage, validation.NilOrNotEmpty, is.URL),
		validation.Field(&u.GroupID, validation.NilOrNotEmpty, is.UUID),
	)
}

type AddAuthProvider struct {
	Provider string  `json:"provider"`
	Password *string `json:"password"`
}

func (a *AddAuthProvider) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.Provider, validation.Required, validation.In("credential", "google").Error("must be one of 'credential', 'google'")),
	)
}

type DeleteManyUser struct {
	IDs []string `json:"ids"`
}

func (u *DeleteManyUser) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.IDs, validation.Required, validation.Length(1, 0), validation.Each(is.UUID)),
	)
}

type CreateManyUsers struct {
	Users []CreateMultiTypeUser `json:"users"`
}

func (c *CreateManyUsers) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Users, validation.Required, validation.Length(1, 0), validation.Each(validation.By(func(value any) error {
			if user, ok := value.(CreateMultiTypeUser); ok {
				return user.Validate()
			}
			return nil
		}))),
	)
}
