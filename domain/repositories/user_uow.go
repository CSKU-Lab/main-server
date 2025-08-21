package repositories

type UserUoWInstance interface {
	User() User
	UserPassword() UserPasswordRepository
	UserGroup() UserGroupRepository
}

type UserUoWRepository interface {
	Execute(cb func(u UserUoWInstance) error) error
}
