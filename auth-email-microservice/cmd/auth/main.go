package main

// 4:07:37

import (
	"log"
	"os"

	"github.com/Chuuch/ecom-microservices/config"
	"github.com/Chuuch/ecom-microservices/internal/server"
	"github.com/Chuuch/ecom-microservices/pkg/logger"
	"github.com/Chuuch/ecom-microservices/pkg/mailer"
	"github.com/Chuuch/ecom-microservices/pkg/postgres"
	"github.com/Chuuch/ecom-microservices/pkg/rabbitmq"
	"github.com/Chuuch/ecom-microservices/pkg/redis"
	"github.com/Chuuch/ecom-microservices/pkg/utils"

	"github.com/Chuuch/ecom-microservices/pkg/jaeger"
	"github.com/opentracing/opentracing-go"
)

func main() {
	log.Println("Starting the application")

	// Init config
	configPath := utils.GetConfigPath(os.Getenv("config"))
	cfgFile, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("LoadConfig: %v", err)
	}

	cfg, err := config.ParseConfig(cfgFile)
	if err != nil {
		log.Fatalf("ParseConfig: %v", err)
	}

	// Init logger
	appLogger := logger.NewApiLogger(cfg)
	appLogger.InitLogger()

	appLogger.Infof("AppVersion: %s, LogLevel: %s, Mode: %s, SSL: %v", cfg.Server.AppVersion, cfg.Logger.Level, cfg.Server.Mode, cfg.Server.SSL)
	appLogger.Infof("Successfully parsed config: %v", cfg.Server.AppVersion)

	// Init PostgreSQL
	psql, err := postgres.NewPsqlDB(cfg)
	if err != nil {
		appLogger.Fatal("PostgreSQL init: %s", err)
	}

	defer psql.Close()

	// Init Redis
	redisClient := redis.NewRedisClient(cfg)
	defer redisClient.Close()
	appLogger.Info("Redis connected")

	// Init Jaeger
	tracer, closer, err := jaeger.InitJaeger(cfg)

	if err != nil {
		appLogger.Fatalf("cannot create tracer: %v", err)
	}

	appLogger.Info("Jaeger connected")

	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	appLogger.Info("Opentracing connected")

	// Init RabbitMQ
	amqpConn, err := rabbitmq.NewRabbitMQConnection(cfg)
	if err != nil {
		appLogger.Fatalf("cannot create RabbitMQ connection: %v", err)
	}
	defer amqpConn.Close()

	// Initialize mailer
	resendClient := mailer.NewResendClient(cfg)
	appLogger.Info("Mailer connected")

	// Init server
	authServer := server.NewAuthServer(appLogger, cfg, psql, redisClient, amqpConn, resendClient)
	appLogger.Fatal(authServer.Start())
}
