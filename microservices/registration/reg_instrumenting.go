package registration

import (
	"github.com/go-kit/kit/metrics"
	"time"
)

func Metrics(counter metrics.Counter, histogram metrics.Histogram) ServiceMiddleware {
	return func(svc Service) Service {
		return metricsMiddleware{
			svc,
			counter,
			histogram,
		}
	}
}

type metricsMiddleware struct {
	Service
	counter   metrics.Counter
	histogram metrics.Histogram
}

func (mw metricsMiddleware) Registration(username, fullname, email, password string, isDisabled bool) (ok bool, e error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "RegistrationViaHTTP"}
		mw.counter.With(lvs...).Add(1)
		mw.histogram.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())

	ok, e = mw.Service.Registration(username, fullname, email, password, isDisabled)
	return
}

func (mw metricsMiddleware) UsernameValidation(username string) (ok bool, e error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "UsernameValidation"}
		mw.counter.With(lvs...).Add(1)
		mw.histogram.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())
	ok, e = mw.Service.UsernameValidation(username)
	return
}

func (mw metricsMiddleware) EmailValidation(email string) (ok bool, e error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "EmailValidation"}
		mw.counter.With(lvs...).Add(1)
		mw.histogram.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())

	ok, e = mw.Service.EmailValidation(email)
	return
}

func (mw metricsMiddleware) RegServiceHealthCheck() (ok bool) {
	defer func(begin time.Time) {
		var lvs []string = []string{"method", "RegServiceHealthCheck"}
		mw.counter.With(lvs...).Add(1)
		mw.histogram.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())

	ok = mw.Service.RegServiceHealthCheck()
	return
}
