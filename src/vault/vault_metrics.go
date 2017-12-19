package vault

import (
	"github.com/go-kit/kit/metrics"
	"time"
	"context"
)

func Metrics(requestCount metrics.Counter, requestLatency  metrics.Histogram) ServiceMiddleware {
	return func(service Service) Service {
		return metricsMiddleware{
			service,
			requestCount,
			requestLatency,
		}
	}
}

type metricsMiddleware struct {
	Service
	requestCount metrics.Counter
	requestLatency metrics.Histogram
}

func (mw metricsMiddleware) Hash(ctx context.Context, password string) (out string, e error)  {
	defer func(begin time.Time) {
		lvs := []string{"method", "Hash"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	out, err := mw.Service.Hash(ctx,password)
	return out, err
}
