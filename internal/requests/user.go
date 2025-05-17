package requests

type BaseUser struct {
	Username    string   `json:"username" db:"username"`
	DisplayName string   `json:"display_name" db:"display_name"`
	Roles       []string `json:"roles" db:"roles"`
}

type CreateUser struct {
	BaseUser
	Email *string `json:"email" db:"email"`
}

type CreateCredentialUser struct {
	BaseUser
	Password string `json:"password" db:"password"`
}

type UpdateUser struct {
	BaseUser
	Email    *string `json:"email" db:"email"`
	Password *string `json:"password" db:"password"`
}

type DeleteManyUser struct {
	IDs []string `json:"ids" db:"ids"`
}
