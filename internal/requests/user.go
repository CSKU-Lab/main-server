package requests

type BaseUser struct {
	Username    string   `json:"username" db:"username"`
	DisplayName string   `json:"display_name" db:"display_name"`
	Roles       []string `json:"roles" db:"roles"`
}


type CreateCredentialUser struct {
	BaseUser
	Password string `json:"password" db:"password"`
	Group    string `json:"group"`
}

type UpdateUser struct {
	BaseUser
	Email        *string `json:"email" db:"email"`
	Password     *string `json:"password" db:"password"`
	ProfileImage *string `json:"profile_image" db:"profile_image"`
}

type DeleteManyUser struct {
	IDs []string `json:"ids" db:"ids"`
}

type CreateMultiTypeUser struct {
	BaseUser
	Type     string  `json:"type" db:"type"`
	Email    *string `json:"email,omitempty" db:"email"`
	Password *string `json:"password,omitempty" db:"password"`
	Group    *string `json:"group,omitempty" db:"group"`
}

type CreateManyUsers struct {
	Users []CreateMultiTypeUser `json:"users" db:"users"`
}
