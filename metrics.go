package main

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registry                           = prometheus.NewRegistry()
	summaryVecExecutionTimeMillisecond *prometheus.SummaryVec
)

func init() {
	metricsNamespace := "cron_runner"
	summaryVecExecutionTimeMillisecond = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: metricsNamespace,
		Name:      "execution_millisecond",
	}, []string{"job_name", "succeeded"})
	registry.MustRegister(summaryVecExecutionTimeMillisecond)
}

type contextKeyMetrics struct{}
type contextValueMetrics struct{ start time.Time }

var ctxKeyMetrics = contextKeyMetrics{}

func startMeasurement(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKeyMetrics, &contextValueMetrics{time.Now()})
}

func finishMeasurement(ctx context.Context, succeeded bool) {
	v := ctx.Value(ctxKeyMetrics).(*contextValueMetrics)
	summaryVecExecutionTimeMillisecond.WithLabelValues(config.jobName,
		fmt.Sprintf("%v", succeeded)).Observe(float64(time.Now().Sub(v.start) / time.Millisecond))
}
