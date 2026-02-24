package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/chuuch/search-microservice/pkg/middlewares"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (a *App) runMetrics(cancel context.CancelFunc) {
	a.metricsServer = echo.New()
	a.metricsServer.Use(
		middleware.RecoverWithConfig(
			middleware.RecoverConfig{
				StackSize:         stackSize,
				DisableStackAll:   false,
				DisablePrintStack: false,
			},
		),
	)
	a.metricsServer.GET(a.cfg.Probes.PrometheusPath, echo.WrapHandler(promhttp.Handler()))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", a.cfg.Probes.PrometheusPort),
		Handler: a.metricsServer,
	}
	a.metricsHttpServer = srv

	go func() {
		a.log.Info("metrics server started on port %s", a.cfg.Probes.PrometheusPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("failed to start metrics server: %v", err)
			cancel()
		}
	}()
}

func (a *App) getHttpMetricsCb() middlewares.MiddlewareMetricsCb {
	return func(err error) {
		if err != nil {
			a.metrics.ErrorHttpRequestsTotal.Inc()
		} else {
			a.metrics.SuccessHttpRequestsTotal.Inc()
		}
	}
}
