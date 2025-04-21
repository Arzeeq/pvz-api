package middleware

import (
	"net/http"
	"time"

	"github.com/Arzeeq/pvz-api/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()

		metrics.HttpRequestsTotal.With(prometheus.Labels{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Inc()

		metrics.HttpRequestDuration.With(prometheus.Labels{
			"method": r.Method,
			"path":   r.URL.Path,
		}).Observe(duration)
	})
}
