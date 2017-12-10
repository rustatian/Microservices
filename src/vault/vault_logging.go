package vault

import (
	"context"
	"github.com/go-kit/kit/log"
	"time"
)

func LoggingMiddleware(logger log.Logger) ServiceMiddleware{
	return func(next Service) Service {
		return loggingMiddleware{next, logger}
	}
}

type loggingMiddleware struct {
	Service
	logger log.Logger
}

func (mv loggingMiddleware) Hash(ctx context.Context, password string) (string, error) {
	defer func(begin time.Time) {
		mv.logger.Log(
			"function", "Hash",
			"password", password,
			"error", error.Error,
		)
	}(time.Now())
	password, err := mv.Service.Hash(ctx, password)
	return password, err
}

