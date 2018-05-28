package vault

import (
	"context"
	"time"

	"go.uber.org/zap"
)

func NewLoggingService(log *zap.SugaredLogger, s Service) Service {
	return &loggingService{
		s,
		log,
	}
}

func (s *loggingService) Hash(ctx context.Context, password string) (string, error) {
	defer s.log.Sync()
	defer func(begin time.Time) {
		s.log.Infow("hash request",
			"ts",     time.Since(begin),
			"password: ", password,
		)
	}(time.Now())
	return s.Service.Hash(ctx, password)
}

func (s *loggingService) Validate(ctx context.Context, password, hash string) (bool, error) {
	defer s.log.Sync()
	defer func(begin time.Time) {
		s.log.Infow("hash request",
			"ts",     time.Since(begin),
			"password: ", password,
			"hash: ",     hash,
		)
	}(time.Now())
	return s.Service.Validate(ctx, password, hash)
}

func (s *loggingService) HealthCheck() bool {
	defer s.log.Sync()
	defer func(begin time.Time) {
		s.log.Infow("hash request",
			"ts",     time.Since(begin),
		)
	}(time.Now())
	return s.Service.HealthCheck()
}

type loggingService struct {
	Service
	log *zap.SugaredLogger
}
