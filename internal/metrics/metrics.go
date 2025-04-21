package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// technical metrics
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path"})

	HttpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_in_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"method", "path"})

	// bussiness metrics
	PvzCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "pvz_created_total",
		Help: "Total number of PVZ created",
	})

	ReceptionsCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "receipts_created_total",
		Help: "Total number of receipts created",
	})

	ProductsAddedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_added_total",
		Help: "Total number of products added",
	})
)
