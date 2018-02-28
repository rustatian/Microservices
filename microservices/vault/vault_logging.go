package vault

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

func (s *loggingService) Hash(ctx context.Context, password string) (string, error) {
	defer func(begin time.Time) {
		s.log.WithFields(logrus.Fields{
			"ts":       time.Since(begin),
			"password": password,
		}).Info("hash request")
	}(time.Now())
	return s.Service.Hash(ctx, password)
}

func (s *loggingService) Validate(ctx context.Context, password, hash string) (bool, error) {
	defer func(begin time.Time) {
		s.log.WithFields(logrus.Fields{
			"ts":         time.Since(begin),
			"password: ": password,
			"hash: ":     hash,
		}).Info("validate request")
	}(time.Now())

	return s.Service.Validate(ctx, password, hash)
}

func (s *loggingService) HealthCheck() bool {
	defer func(begin time.Time) {
		defer func() {
			s.log.WithFields(logrus.Fields{
				"ts": time.Since(begin),
			}).Info("health request")
		}()
	}(time.Now())
	return s.Service.HealthCheck()
}

type loggingService struct {
	Service
	log *logrus.Logger
}
