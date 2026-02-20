package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Chuuch/ecom-microservices/config"
	"github.com/Chuuch/ecom-microservices/internal/interceptors"
	"github.com/Chuuch/ecom-microservices/pkg/logger"
	"github.com/Chuuch/ecom-microservices/pkg/metric"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/resend/resend-go/v2"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	sessionRepository "github.com/Chuuch/ecom-microservices/internal/session/repository"
	sessionUseCase "github.com/Chuuch/ecom-microservices/internal/session/usecase"
	authServerGRPC "github.com/Chuuch/ecom-microservices/internal/user/delivery/grpc/service"
	"github.com/Chuuch/ecom-microservices/internal/user/repository"
	"github.com/Chuuch/ecom-microservices/internal/user/usecase"
	userService "github.com/Chuuch/ecom-microservices/proto"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"

	emailServerGRPC "github.com/Chuuch/ecom-microservices/internal/email/delivery/grpc"

	"github.com/Chuuch/ecom-microservices/internal/email/delivery/rabbitmq"
	"github.com/Chuuch/ecom-microservices/internal/email/mailer"
	emailRepository "github.com/Chuuch/ecom-microservices/internal/email/repository"
	emailUseCase "github.com/Chuuch/ecom-microservices/internal/email/usecase"
)

// GRPC Auth Server
type Server struct {
	logger       logger.Logger
	cfg          *config.Config
	db           *sqlx.DB
	redis        *redis.Client
	amqpConn     *amqp.Connection
	resendClient *resend.Client
}

func NewAuthServer(logger logger.Logger, cfg *config.Config, db *sqlx.DB, redis *redis.Client, amqpConn *amqp.Connection, resendClient *resend.Client) *Server {
	return &Server{
		logger:       logger,
		cfg:          cfg,
		db:           db,
		redis:        redis,
		amqpConn:     amqpConn,
		resendClient: resendClient,
	}
}

func (s *Server) Start() error {

	// Metrics
	metrics, err := metric.CreateMetrics(s.cfg.Metric.URL, s.cfg.Metric.ServiceName)
	if err != nil {
		s.logger.Errorf("metric.CreateMetrics: %v", err)
	}

	im := interceptors.NewInterceptorManager(s.logger, s.cfg, metrics)
	s.logger.Info("Metrics available URL: %s", s.cfg.Metric.URL)

	// User
	userPGRepo := repository.NewUserPGRepository(s.db)
	userRedisRepo := repository.NewUserRedisRepository(s.redis, s.logger)
	userUC := usecase.NewUserUseCase(s.logger, userPGRepo, userRedisRepo)

	// Session
	sessionRepo := sessionRepository.NewSessionRepository(s.redis, s.cfg)
	sessionUC := sessionUseCase.NewSessionUseCase(sessionRepo, s.cfg)

	// Email
	emailRepo := emailRepository.NewEmailRepository(s.db)
	mailDialer := mailer.NewMailer(s.resendClient)

	emailsPublisher, err := rabbitmq.NewEmailsPublisher(s.cfg, s.logger)
	if err != nil {
		return err
	}
	defer emailsPublisher.CloseMessagesChannel()
	s.logger.Info("Emails publisher initialized")

	emailUC := emailUseCase.NewEmailUseCase(emailRepo, s.logger, mailDialer, s.cfg, emailsPublisher)
	emailAmqpConsumer := rabbitmq.NewEmailConsumer(s.amqpConn, s.logger, emailUC)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := emailAmqpConsumer.StartConsumer(
			s.cfg.RabbitMQ.WorkerPoolSize,
			s.cfg.RabbitMQ.Exchange,
			s.cfg.RabbitMQ.Queue,
			s.cfg.RabbitMQ.RoutingKey,
			s.cfg.RabbitMQ.ConsumerTag,
		)
		if err != nil {
			s.logger.Errorf("emailAmqpConsumer.StartConsumer: %v", err)
		}
	}()

	l, err := net.Listen("tcp", s.cfg.Server.Port)
	if err != nil {
		return err
	}

	defer l.Close()

	server := grpc.NewServer(grpc.KeepaliveParams(
		keepalive.ServerParameters{
			MaxConnectionIdle: s.cfg.Server.MaxConnectionIdle * time.Minute,
			Timeout:           s.cfg.Server.Timeout * time.Second,
			MaxConnectionAge:  s.cfg.Server.MaxConnectionAge * time.Minute,
			Time:              s.cfg.Server.Time * time.Minute,
		},
	), grpc.ChainUnaryInterceptor(im.Logger, im.Metrics, grpcrecovery.UnaryServerInterceptor(), grpcPrometheus.UnaryServerInterceptor))

	if s.cfg.Server.Mode != "Production" {
		reflection.Register(server)
	}

	authGRPCServer := authServerGRPC.NewAuthServerGRPC(s.logger, s.cfg, userUC, sessionUC, metrics, emailUC)
	emailGRPCServer := emailServerGRPC.NewEmailMicroservice(emailUC, s.cfg, s.logger)

	userService.RegisterUserServiceServer(server, authGRPCServer)
	userService.RegisterEmailServiceServer(server, emailGRPCServer)

	grpcPrometheus.Register(server)
	http.Handle("/metrics", promhttp.Handler())

	s.logger.Info("Server is listening on port: %v", s.cfg.Server.Port)
	if err := server.Serve(l); err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	select {
	case v := <-quit:
		s.logger.Errorf("Received signal: %v", v)
	case done := <-ctx.Done():
		s.logger.Errorf("Received context done: %v", done)
	}
	server.GracefulStop()
	s.logger.Info("Shutting down server...")

	return nil
}
