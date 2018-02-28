package authorization

import (
	"github.com/go-kit/kit/metrics"
	"time"
)

func Metrics(requestCount metrics.Counter, requestLatency metrics.Histogram) ServiceMiddleware {
	return func(svc Service) Service {
		return metricsMiddleware{
			svc,
			requestCount,
			requestLatency,
		}
	}
}

type metricsMiddleware struct {
	Service
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
}

func (mw metricsMiddleware) Login(username, password string) (mesg string, roles []string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Login"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())

	mesg, roles, err = mw.Service.Login(username, password)
	return
}

func (mw metricsMiddleware) Logout() (out string) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Logout"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())

	out = mw.Service.Logout()
	return
}

func (mw metricsMiddleware) AuthHealtCheck() (res bool) {
	defer func(begin time.Time) {
		lvs := []string{"method", "AuthHealtCheck"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())

	res = mw.Service.AuthHealtCheck()
	return
}
