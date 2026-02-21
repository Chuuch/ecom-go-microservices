package app

import (
	"context"
	"fmt"
	"io"
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
	"github.com/pkg/errors"
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

func (a *App) initIndexes(ctx context.Context) error {
	exists, err := a.isIndexExists(ctx, a.cfg.ElasticMapping.Name)
	if err != nil {
		return err
	}

	if !exists {
		if err := a.uploadElasticMappings(ctx, a.cfg.ElasticMapping.ProductsIndex); err != nil {
			return err
		}
	}

	a.log.Infof("Index %s created", a.cfg.ElasticMapping.Name)
	return nil
}

func (a *App) isIndexExists(ctx context.Context, indexName string) (bool, error) {
	response, err := esclient.Exists(ctx, a.elasticClient, []string{indexName})
	if err != nil {
		a.log.Errorf("Failed to check if index exists: %v", err)
		return false, errors.Wrap(err, "failed to check if index exists")
	}

	defer response.Body.Close()

	a.log.Info("Index exists response: %s", response.String())

	exists := response.StatusCode == 200

	return exists, nil
}

func (a *App) uploadElasticMappings(ctx context.Context, indexConfig esclient.ElasticIndex) error {
	getwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get working directory")
	}

	path := fmt.Sprintf("%s/%s", getwd, &indexConfig.Path)

	mappingsFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer mappingsFile.Close()
	
	mappingBytes, err := io.ReadAll(mappingsFile)
	if err != nil {
		return err
	}

	a.log.Info("loaded elastic mappings file %s", path)

	response, err := esclient.CreateIndex(ctx, a.elasticClient, indexConfig.Name, mappingBytes)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.IsError() {
		return errors.New(response.String())
	}

	a.log.Info("created index %s", response.String())
	return nil

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

	if err := a.initIndexes(ctx); err != nil {
		return err
	}

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
