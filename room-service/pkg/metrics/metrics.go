package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (

	// Quantidade de requisições HTTP por rota + request + status
	HTTPRequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Número total de requisições HTTP",
		},
		[]string{"room-service", "method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duração das requisições HTTP",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"room-service", "method", "path"},
	)
)
