package vault

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
)

func NewLoggingService(log log.Logger, s Service) Service {
	return &loggingService{
		s,
		log,
	}
}

func (s *loggingService) Hash(ctx context.Context, password string) (string, error) {
	defer func(timer time.Time) {
		s.log.Log(
			"method", "hash",
			"password: ", password,
		)
	}(time.Now())
	return s.Service.Hash(ctx, password)
}

func (s *loggingService) Validate(ctx context.Context, password, hash string) (bool, error) {
	defer func(timer time.Time) {
		s.log.Log(
			"method", "validate",
			"password: ", password,
			"hash: ", hash,
		)
	}(time.Now())

	return s.Service.Validate(ctx, password, hash)
}

func (s *loggingService) HealthCheck() bool {
	defer func(begin time.Time) {
		defer func() {
			s.log.Log(
				"method", "health",
			)
		}()
	}(time.Now())
	return s.Service.HealthCheck()
}

//func LoggingMiddleware(logger log.Logger) endpoint.Middleware {
//	return func(next endpoint.Endpoint) endpoint.Endpoint {
//		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
//			defer func(begin time.Time) {
//				//logger.Log("transport_error", err, "took", time.Since(begin))
//				switch v := request.(type) {
//				case healthRequest:
//					logger.Log("transport_error", err, "took", time.Since(begin), "health", "healthReq")
//				case hashRequest:
//					logger.Log("transport_error", err, "took", time.Since(begin), "pass", v.Password)
//				default:
//					logger.Log("transport_error", err, "took", time.Since(begin))
//				}
//			}(time.Now())
//			return next(ctx, request)
//		}
//	}
//}

type loggingService struct {
	Service
	log log.Logger
}
