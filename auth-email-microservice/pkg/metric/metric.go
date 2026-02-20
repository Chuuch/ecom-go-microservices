package metric

import (
	"log"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics interface {
	IncHits(status int, method, path string)
	ObserveResponseTime(status int, method, path string, observeTime float64)
}

type PrometheusMetrics struct {
	HitsTotal prometheus.Counter
	Hits      *prometheus.CounterVec
	Times     *prometheus.HistogramVec
}

func CreateMetrics(address, name string) (Metrics, error) {
	var metrics PrometheusMetrics
	metrics.HitsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: name + "_hits_total",
	})

	if err := prometheus.Register(metrics.HitsTotal); err != nil {
		return nil, err
	}

	metrics.Hits = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name + "_hits",
	}, []string{"status", "method", "path"})

	if err := prometheus.Register(metrics.Hits); err != nil {
		return nil, err
	}

	metrics.Times = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: name + "_times",
	}, []string{"status", "method", "path"})

	if err := prometheus.Register(metrics.Times); err != nil {
		return nil, err
	}

	if err := prometheus.Register(collectors.NewBuildInfoCollector()); err != nil {
		return nil, err
	}

	go func() {
		router := echo.New()
		router.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
		if err := router.Start(address); err != nil {
			log.Fatal(err)
		}
	}()

	return &metrics, nil
}

func (m *PrometheusMetrics) IncHits(status int, method, path string) {
	m.HitsTotal.Inc()
	m.Hits.WithLabelValues(strconv.Itoa(status), method, path).Inc()
}

func (m *PrometheusMetrics) ObserveResponseTime(status int, method, path string, observeTime float64) {
	m.Times.WithLabelValues(strconv.Itoa(status), method, path).Observe(observeTime)
}
