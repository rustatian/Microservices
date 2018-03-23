package registration

import . "github.com/ValeryPiashchynski/TaskManager/microservices/registration/application"

type Repository interface {
	Registration(user *User) (bool, error)
	UsernameValidation(user *User) (bool, error)
	EmailValidation(user *User) (bool, error)
}
