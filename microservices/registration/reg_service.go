package registration

import (
	"context"

	"github.com/ValeryPiashchynski/TaskManager/microservices/registration/application"
	"github.com/ValeryPiashchynski/TaskManager/microservices/registration/registration"
)

type Service interface {
	Registration(ctx context.Context, username, fullname, email, password string, isDisabled bool) (bool, error)
	UsernameValidation(ctx context.Context, username string) (bool, error)
	EmailValidation(ctx context.Context, email string) (bool, error)
	RegServiceHealthCheck() bool
}

func NewRegService(repository registration.Repository) Service {
	return &newRegService{
		rep: repository,
	}
}

type newRegService struct {
	rep registration.Repository
}

type ServiceMiddleware func(svc Service) Service

func (r *newRegService) Registration(ctx context.Context, username, fullname, email, password string, isDisabled bool) (ok bool, e error) {
	u := application.User{
		Username:   username,
		FullName:   fullname,
		Email:      email,
		Password:   password,
		IsDisabled: isDisabled,
	}

	res, err := r.rep.Registration(&u)
	if err != nil {
		return res, err
	}
	return res, nil
}

func (r *newRegService) UsernameValidation(ctx context.Context, username string) (bool, error) {
	u := application.User{
		Username: username,
	}

	res, err := r.rep.UsernameValidation(&u)
	if err != nil {
		return res, err
	}
	return res, nil
}

func (r *newRegService) EmailValidation(ctx context.Context, email string) (bool, error) {
	u := application.User{
		Email: email,
	}

	res, err := r.rep.EmailValidation(&u)
	if err != nil {
		return res, err
	}
	return res, nil
}

func (newRegService) RegServiceHealthCheck() bool {
	return true
}
