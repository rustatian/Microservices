package calendar

import (
	"context"
	"github.com/go-kit/kit/metrics"
	"time"
)

func Metrics(requestCount metrics.Counter, requestLatency metrics.Histogram) ServiceMiddleware {
	return func(svc TaskService) TaskService {
		return metricsMiddleware{
			svc,
			requestCount,
			requestLatency,
		}
	}
}

type metricsMiddleware struct {
	TaskService
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
}

func (mw metricsMiddleware) GetTasks(ctx context.Context, username string, tr timeRange) (resp string, er error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "GetTasks"}
		mw.requestCount.With(lvs...).Add(1)
		mw.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds() * 100000)
	}(time.Now())

	resp, er = mw.TaskService.GetTasks(ctx, username, tr)
	return
}
