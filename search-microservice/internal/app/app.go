package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chuuch/search-microservice/config"
	"github.com/chuuch/search-microservice/pkg/elastic"
	"github.com/chuuch/search-microservice/pkg/esclient"
	"github.com/chuuch/search-microservice/pkg/logger"
	"github.com/chuuch/search-microservice/pkg/logger/jaeger"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-playground/validator"
	"github.com/opentracing/opentracing-go"
)

type App struct {
	log           logger.Logger
	cfg           *config.Config
	doneCh        chan struct{}
	elasticClient *elasticsearch.Client
	validate      *validator.Validate
}

func NewApp(log logger.Logger, cfg *config.Config) *App {
	return &App{
		log:      log,
		cfg:      cfg,
		doneCh:   make(chan struct{}),
		validate: validator.New(),
	}
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Initialize Jaeger
	if a.cfg.Jaeger.Enable {
		tracer, closer, err := jaeger.InitJaeger(a.cfg)
		if err != nil {
			a.log.Fatalf("Failed to initialize Jaeger: %v", err)
		}
		defer closer.Close()
		opentracing.SetGlobalTracer(tracer)
		a.log.Info("Jaeger initialized")
	}

	// Initialize Elasticsearch
	elasticSearchClient, err := elastic.NewElasticSearch(a.cfg)
	if err != nil {
		return err
	}
	a.elasticClient = elasticSearchClient
	a.log.Info("ElasticSearch client initialized")

	// Check if ElasticSearch is reachable
	elasticResponse, err := esclient.Info(ctx, a.elasticClient)
	if err != nil {
		return err
	}

	a.log.Infof("ElasticSearch is reachable %s", elasticResponse.String())

	<-ctx.Done()
	a.waitShutDown(3 * time.Second)

	<-a.doneCh
	a.log.Info("App exited properly")
	return nil
}

func (a *App) waitShutDown(duration time.Duration) {
	go func() {
		time.Sleep(duration)
		a.doneCh <- struct{}{}
	}()
}
