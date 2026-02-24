package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/chuuch/search-microservice/pkg/esclient"
	"github.com/heptiolabs/healthcheck"
)

func (a *App) runHealthCheck(ctx context.Context) {
	health := healthcheck.NewHandler()

	mux := http.NewServeMux()
	mux.HandleFunc(a.cfg.Probes.ReadinessPath, health.ReadyEndpoint)
	mux.HandleFunc(a.cfg.Probes.LivenessPath, health.LiveEndpoint)

	a.healthcheckServer = &http.Server{
		Handler:      mux,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
		Addr:         fmt.Sprintf(":%s", a.cfg.Probes.Port),
	}

	a.configureHealthCheckEndpoints(ctx, health)

	go func() {
		a.log.Info("health check server started on port %s", a.cfg.Probes.Port)
		if err := a.healthcheckServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("failed to start health check server: %v", err)
		}
	}()
}

func (a *App) configureHealthCheckEndpoints(ctx context.Context, health healthcheck.Handler) {
	health.AddReadinessCheck("elasticsearch", healthcheck.AsyncWithContext(ctx, func() error {
		_, err := esclient.Info(ctx, a.elasticClient)
		if err != nil {
			a.log.Error("failed to check elasticsearch: %v", err)
			return err
		}
		return nil
	}, time.Duration(a.cfg.Probes.CheckIntervalSeconds)*time.Second))

	health.AddLivenessCheck("elasticsearch", healthcheck.AsyncWithContext(ctx, func() error {
		_, err := esclient.Info(ctx, a.elasticClient)
		if err != nil {
			a.log.Error("failed to check elasticsearch: %v", err)
			return err
		}
		return nil
	}, time.Duration(a.cfg.Probes.CheckIntervalSeconds)*time.Second))

	health.AddLivenessCheck("rabbitmq", healthcheck.AsyncWithContext(ctx, func() error {
		if a.amqpConn.IsClosed() || a.amqpChan.IsClosed() {
			a.log.Error("rabbitmq is not connected")
			return errors.New("rabbit mq is not connected")
		}
		return nil
	}, time.Duration(a.cfg.Probes.CheckIntervalSeconds)*time.Second))
}

func (a *App) shutDownHealthCheckServer(ctx context.Context) error {
	return a.healthcheckServer.Shutdown(ctx)
}
