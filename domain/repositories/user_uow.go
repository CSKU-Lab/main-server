package repositories

type UserUoWInstance interface {
	User() UserRepository
	UserPassword() UserPasswordRepository
	UserGroup() UserGroupRepository
}

type UserUoWRepository interface {
	Execute(cb func(u UserUoWInstance) error) error
}
