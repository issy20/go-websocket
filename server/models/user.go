package models

type IUser interface {
	GetId() string
	GetName() string
}

type UserRepository interface {
	AddUser(id string, name string, username string, password string) error
	RemoveUser(user IUser)
	FindUserById(ID string) IUser
	GetAllUsers() []IUser
}
