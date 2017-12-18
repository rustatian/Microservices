package auth

import (
	"context"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// implement function to return ServiceMiddleware
func LoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(i endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				logger.Log("transport_error", err, "took", time.Since(begin))
			}(time.Now())
			return i(ctx, request)
		}
	}
}

type loggingMiddleware struct {
	Service
	logger log.Logger
}

// Implement Service Interface for LoggingMiddleware
func (mw loggingMiddleware) Login(username string, password string) (mesg string, roles []string, err error) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "Login",
			"mesg", mesg,
			"roles", strings.Join(roles, ","),
			"took", time.Since(begin),
		)
	}(time.Now())
	mesg, roles, err = mw.Service.Login(username, password)
	return
}

func (mw loggingMiddleware) Logout() (mesg string) {
	defer func(begin time.Time) {
		mw.logger.Log(
			"function", "Logout",
			"result", mesg,
			"took", time.Since(begin),
		)
	}(time.Now())
	mesg = mw.Service.Logout()
	return
}
