package middleware

import (
	"github.com/chuuch/product-microservice/config"
	"github.com/chuuch/product-microservice/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpTotalRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_service_http_total_requests",
		Help: "Total number of incoming http requests",
	})
)

// MiddlewareManager http midlewares
type middlewareManager struct {
	log logger.Logger
	cfg *config.Config
}

// MiddelwareManager interface
type MiddlewareManager interface {
	Metrics(next echo.HandlerFunc) echo.HandlerFunc
}

// NewMiddlewareManager constructor
func NewMiddlewareManager(log logger.Logger, cfg *config.Config) *middlewareManager {
	return &middlewareManager{
		log: log,
		cfg: cfg,
	}
}

// Metrics prometheus metrics
func (m *middlewareManager) Metrics(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		httpTotalRequests.Inc()
		return next(c)
	}
}
