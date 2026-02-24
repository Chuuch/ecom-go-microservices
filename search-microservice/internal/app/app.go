package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chuuch/search-microservice/config"
	"github.com/chuuch/search-microservice/internal/metrics"
	"github.com/chuuch/search-microservice/internal/product/repository"
	rabbitmqConsumer "github.com/chuuch/search-microservice/internal/product/transport/http/rabbitmq"
	v1 "github.com/chuuch/search-microservice/internal/product/transport/http/v1"
	"github.com/chuuch/search-microservice/internal/product/usecase"
	"github.com/chuuch/search-microservice/pkg/esclient"
	"github.com/chuuch/search-microservice/pkg/jaeger"
	"github.com/chuuch/search-microservice/pkg/logger"
	"github.com/chuuch/search-microservice/pkg/middlewares"
	misstypemanager "github.com/chuuch/search-microservice/pkg/misstype_manager"
	"github.com/chuuch/search-microservice/pkg/rabbitmq"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v5"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

type App struct {
	log               logger.Logger
	cfg               *config.Config
	doneCh            chan struct{}
	elasticClient     *elasticsearch.Client
	validate          *validator.Validate
	echo              *echo.Echo
	httpServer        *http.Server
	middlewareManager middlewares.MiddlewareManager
	misstypeManager   misstypemanager.MisstypeManager
	amqpConn          *amqp.Connection
	amqpChan          *amqp.Channel
	amqpConsumeChan   *amqp.Channel
	amqpPublisher     rabbitmq.AmqpPublisher
	metrics           *metrics.SearchMicroserviceMetrics
	metricsServer     *echo.Echo
	metricsHttpServer *http.Server
	healthcheckServer *http.Server
}

func (a *App) loadKeysMappings() (*misstypemanager.KeyboardMisstypeManager, error) {
	getwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get working directory")
	}

	keysJsonPath := fmt.Sprintf("%s/config/translate.json", getwd)
	keysJsonPathFile, err := os.Open(keysJsonPath)
	if err != nil {
		return nil, err
	}

	defer keysJsonPathFile.Close()

	keysJsonBytes, err := io.ReadAll(keysJsonPathFile)
	if err != nil {
		return nil, err
	}

	a.log.Info("loaded keys mappings file %s", keysJsonPath)

	keyMappings := map[string]string{}
	if err := json.Unmarshal(keysJsonBytes, &keyMappings); err != nil {
		return nil, err
	}

	a.log.Info("loaded %d key mappings", len(keyMappings))

	misstypeManager := misstypemanager.NewMisstypeManager(a.log, keyMappings)
	return misstypeManager, nil
}

func NewApp(log logger.Logger, cfg *config.Config) *App {
	return &App{
		log:      log,
		cfg:      cfg,
		doneCh:   make(chan struct{}),
		validate: validator.New(),
		echo:     echo.New(),
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

	path := fmt.Sprintf("%s/%s", getwd, indexConfig.Path)

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

	misstypeManager, err := a.loadKeysMappings()
	if err != nil {
		return err
	}
	a.misstypeManager = misstypeManager
	a.log.Info("misstype manager initialized")

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

	a.middlewareManager = middlewares.NewMiddlewareManager(a.log, a.cfg, a.getHttpMetricsCb())

	a.metrics = metrics.NewSearchMicroserviceMetrics(a.cfg)

	// Initialize RabbitMQ
	if err := a.initRabbitMQ(ctx); err != nil {
		return err
	}
	defer a.closeRabbitMQ()

	queue, err := rabbitmq.DeclareBinding(ctx, a.amqpChan, rabbitmq.ExchangeAndQueueBinding{
		ExchangeName: a.cfg.RabbitMQ.ExchangeName,
		ExchangeKind: a.cfg.RabbitMQ.ExchangeKind,
		QueueName:    a.cfg.RabbitMQ.QueueName,
		BindingKey:   a.cfg.RabbitMQ.BindingKey,
	})
	if err != nil {
		return err
	}
	a.log.Info("RabbitMQ queue declared: %s", queue.Name)

	consumeChan, err := a.amqpConn.Channel()
	if err != nil {
		a.log.Error("Failed to open consume channel: %v", err)
		return err
	}
	if err := consumeChan.Qos(1, 0, true); err != nil {
		a.log.Error("failed to set QoS on consume channel: %v", err)
		_ = consumeChan.Close()
		return err
	}
	a.amqpConsumeChan = consumeChan

	if err := a.initRabbitMQPublisher(ctx); err != nil {
		a.log.Error("Failed to initialize RabbitMQ publisher: %v", err)
		return err
	}

	// Initialize Elasticsearch
	if err := a.initElasticSearchClient(ctx); err != nil {
		a.log.Error("Failed to initialize ElasticSearch client: %v", err)
		return err
	}

	if err := a.initIndexes(ctx); err != nil {
		return err
	}

	elasticRepository := repository.NewElasticRepository(a.log, a.cfg, a.elasticClient, a.misstypeManager)
	productUsecase := usecase.NewProductUsecase(a.log, a.cfg, elasticRepository, a.amqpPublisher)
	productController := v1.NewProductController(a.log, a.cfg, productUsecase, a.echo.Group(a.cfg.Http.ProductsPath), a.validate)
	productController.MapRoutes()

	go func() {
		if err := a.runHttpServer(); err != nil {
			a.log.Error("Failed to start HTTP server: %v", err)
			cancel()
		}
	}()

	a.log.Info("HTTP server started on port %s", a.cfg.Http.Port)

	productConsumer := rabbitmqConsumer.NewProductConsumer(a.log, a.cfg, a.amqpConn, a.amqpConsumeChan, productUsecase, a.elasticClient)
	if err := productConsumer.InitBulkIndexer(); err != nil {
		a.log.Error("Failed to initialize bulk indexer: %v", err)
		cancel()
	}
	defer productConsumer.Close(ctx)

	go func() {
		if err := rabbitmq.ConsumeQueue(
			ctx,
			a.amqpConsumeChan,
			a.cfg.RabbitMQ.Concurrency,
			queue.Name,
			a.cfg.RabbitMQ.Consumer,
			productConsumer.ConsumeIndexDeliveries,
		); err != nil {
			a.log.Error("Failed to consume queue: %v", err)
			cancel()
		}
	}()

	a.runMetrics(cancel)
	a.runHealthCheck(ctx)

	<-ctx.Done()
	a.waitShutDown(3 * time.Second)

	if a.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			a.log.Error("Failed to shutdow HTTP server: %v", err)
			return err
		}
	}

	if a.metricsHttpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := a.metricsHttpServer.Shutdown(shutdownCtx); err != nil {
			a.log.Error("Failed to shutdown metrics server: %v", err)
			return err
		}
	}

	if err := a.shutDownHealthCheckServer(ctx); err != nil {
		a.log.Error("Failed to shutdown health check server: %v", err)
		return err
	}

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
