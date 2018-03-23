package vault

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
)

func NewInstrumentingService(requestCount metrics.Counter, requestLatency metrics.Histogram, s Service) Service {
	return &metricsMiddleware{
		requestCount:   requestCount,
		requestLatency: requestLatency,
		Service:        s,
	}
}

type metricsMiddleware struct {
	Service
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
}

func (mw metricsMiddleware) Hash(ctx context.Context, password string) (out string, e error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Hash"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())
	out, e = mw.Service.Hash(ctx, password)
	return
}

func (mw metricsMiddleware) Validate(ctx context.Context, password, hash string) (out bool, e error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Validate"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())
	out, e = mw.Service.Validate(ctx, password, hash)
	return
}

func (mw metricsMiddleware) HealthCheck() (res bool) {
	defer func(begin time.Time) {
		lvs := []string{"method", "HealthCheck"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())
	res = mw.Service.HealthCheck()
	return
}
