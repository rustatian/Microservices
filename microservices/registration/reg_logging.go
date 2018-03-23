package registration

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

func NewLoggingService(log *logrus.Logger, s Service) Service {
	return &loggingService{
		s,
		log,
	}
}

func (s *loggingService) Registration(ctx context.Context, username, fullname, email, password string, isDisabled bool) (bool, error) {
	defer func(begin time.Time) {
		s.log.WithFields(logrus.Fields{
			"ts":       time.Since(begin),
			"username": username,
			"fullname": fullname,
			"email":    email,
			"password": password,
		}).Info("registration request")
	}(time.Now())
	return s.Service.Registration(ctx, username, fullname, email, password, isDisabled)
}

func (s *loggingService) UsernameValidation(ctx context.Context, username string) (bool, error) {
	defer func(begin time.Time) {
		s.log.WithFields(logrus.Fields{
			"ts":       time.Since(begin),
			"username": username,
		}).Info("username validation")
	}(time.Now())
	return s.Service.UsernameValidation(ctx, username)
}

func (s *loggingService) EmailValidation(ctx context.Context, email string) (bool, error) {
	defer func(begin time.Time) {
		s.log.WithFields(logrus.Fields{
			"ts":    time.Since(begin),
			"email": email,
		})
	}(time.Now())
	return s.Service.EmailValidation(ctx, email)
}

func (s *loggingService) RegServiceHealthCheck() bool {
	defer func(begin time.Time) {
		s.log.WithFields(logrus.Fields{
			"ts": time.Since(begin),
		})
	}(time.Now())
	return s.Service.RegServiceHealthCheck()
}

type loggingService struct {
	Service
	log *logrus.Logger
}
