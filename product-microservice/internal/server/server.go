package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chuuch/product-microservice/config"
	"github.com/chuuch/product-microservice/internal/interceptors"
	"github.com/chuuch/product-microservice/internal/product/repository"
	"github.com/chuuch/product-microservice/internal/product/usecase"
	"github.com/chuuch/product-microservice/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/mongo"

	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	productService "github.com/chuuch/product-microservice/internal/product/delivery/gRPC"
	productsService "github.com/chuuch/product-microservice/proto/product"
)

// Server struct
type Server struct {
	logger  logger.Logger
	cfg     *config.Config
	tracer  opentracing.Tracer
	mongoDB *mongo.Client
}

func NewServer(logger logger.Logger, cfg *config.Config, tracer opentracing.Tracer, mongoDB *mongo.Client) *Server {
	return &Server{
		logger:  logger,
		cfg:     cfg,
		tracer:  tracer,
		mongoDB: mongoDB,
	}
}

func (s *Server) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	validate := validator.New()

	productMongoRepo := repository.NewProductMongoRepository(s.mongoDB)
	productUC := usecase.NewProductUC(productMongoRepo, s.logger)

	im := interceptors.NewInterceptorManager(s.logger, s.cfg)

	l, err := net.Listen("tcp", s.cfg.Server.Port)
	if err != nil {
		return errors.Wrap(err, "net.Listen")
	}

	defer l.Close()

	grpcServer := grpc.NewServer(grpc.KeepaliveParams(
		keepalive.ServerParameters{
			MaxConnectionIdle: s.cfg.Server.MaxConnectionIdle * time.Minute,
			MaxConnectionAge:  s.cfg.Server.MaxConnectionAge * time.Minute,
			Timeout:           s.cfg.Server.Timeout * time.Second,
			Time:              s.cfg.Server.Timeout * time.Minute,
		}),
		grpc.ChainUnaryInterceptor(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpcrecovery.UnaryServerInterceptor(),
			im.Logger,
		))

	productService := productService.NewProductGRPCService(productUC, s.logger, validate)
	productsService.RegisterProductServiceServer(grpcServer, productService)
	grpc_prometheus.Register(grpcServer)

	go func() {
		s.logger.Infof("GRPC Server is running on port: %s", s.cfg.Server.Port)
		s.logger.Fatal(grpcServer.Serve(l))
	}()

	if s.cfg.Server.Development {
		reflection.Register(grpcServer)
	}

	metricsServer := echo.New()

	go func() {
		metricsServer.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
		s.logger.Infof("Metrics server is running on port %s", s.cfg.Metrics.Port)
		if err := metricsServer.Start(s.cfg.Metrics.Port); err != nil {
			s.logger.Fatal(err)
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case v := <-quit:
		s.logger.Errorf("signal.Notify: %v", v)
	case done := <-ctx.Done():
		s.logger.Errorf("ctx.Done: %v", done)
	}

	if err := metricsServer.Shutdown(ctx); err != nil {
		s.logger.Errorf("metricsServer.Shutdown: %v", err)
	}

	grpcServer.GracefulStop()
	s.logger.Info("Shutting down gracefully...")

	return nil
}
