package repositories

type UserUoWInstance interface {
	User() User
	UserPassword() UserPassword
	UserGroup() UserGroup
}

type UserUoWRepository interface {
	Execute(cb func(u UserUoWInstance) error) error
}
