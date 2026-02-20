package v1

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	successRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_success_http_requests_total",
		Help: "Total number of successful HTTP requests",
	})

	errorRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_error_http_requests_total",
		Help: "Total number of failed HTTP requests",
	})

	createRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_create_http_requests_total",
		Help: "Total number of create HTTP requests",
	})

	updateRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_update_http_requests_total",
		Help: "Total number of update HTTP requests",
	})

	getRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_get_http_requests_total",
		Help: "Total number of get HTTP requests",
	})

	searchRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_search_http_requests_total",
		Help: "Total number of search HTTP requests",
	})
)
