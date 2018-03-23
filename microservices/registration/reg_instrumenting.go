package registration

import (
	"context"
	"github.com/go-kit/kit/metrics"
	"time"
)

func NewInstrumentingService(counter metrics.Counter, histogram metrics.Histogram, s Service) Service {
	return &metricsMiddleware{
		s,
		counter,
		histogram,
	}
}

type metricsMiddleware struct {
	Service
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
}

func (mw metricsMiddleware) Registration(ctx context.Context, username, fullname, email, password string, isDisabled bool) (ok bool, e error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "RegistrationViaHTTP"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())

	ok, e = mw.Service.Registration(ctx, username, fullname, email, password, isDisabled)
	return
}

func (mw metricsMiddleware) UsernameValidation(ctx context.Context, username string) (ok bool, e error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "UsernameValidation"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())
	ok, e = mw.Service.UsernameValidation(ctx, username)
	return
}

func (mw metricsMiddleware) EmailValidation(ctx context.Context, email string) (ok bool, e error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "EmailValidation"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())

	ok, e = mw.Service.EmailValidation(ctx, email)
	return
}

func (mw metricsMiddleware) RegServiceHealthCheck() (ok bool) {
	defer func(begin time.Time) {
		var lvs []string = []string{"method", "RegServiceHealthCheck"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())

	ok = mw.Service.RegServiceHealthCheck()
	return
}
