package vault

import (
	"context"
	"github.com/go-kit/kit/metrics"
	"time"
)

func Metrics(requestCount metrics.Counter, requestLatency metrics.Histogram, svc Service) Service {
		return metricsMiddleware{
			svc,
			requestCount,
			requestLatency,
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
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	out, e = mw.Service.Hash(ctx, password)
	return
}

func (mw metricsMiddleware) Validate(ctx context.Context, password, hash string) (out bool, e error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "Validate"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	out, e = mw.Service.Validate(ctx, password, hash)
	return
}
