package main

import (
	"context"
	"log"

	"github.com/chuuch/product-microservice/config"
	"github.com/chuuch/product-microservice/internal/server"
	"github.com/chuuch/product-microservice/pkg/jaeger"
	"github.com/chuuch/product-microservice/pkg/kafka"
	"github.com/chuuch/product-microservice/pkg/logger"
	"github.com/chuuch/product-microservice/pkg/mongodb"
	"github.com/chuuch/product-microservice/pkg/redis"
	"github.com/opentracing/opentracing-go"
)

func main() {
	log.Printf("Starting product microservice")

	// Init context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Init config
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("ParseConfig: %v", err)
	}

	// Init logger
	appLogger := logger.NewApiLogger(cfg)
	appLogger.InitLogger()
	appLogger.Info("Logger initialized")
	appLogger.Infof(
		"AppVersion: %s, LogLevel: %s, DevelopmentMode: %s",
		cfg.AppVersion,
		cfg.Logger.Level,
		cfg.Server.Development,
	)
	appLogger.Infof("Successfully parsed config: %v", cfg.AppVersion)

	// Init Jaeger
	tracer, closer, err := jaeger.InitJaeger(cfg)
	if err != nil {
		appLogger.Fatalf("cannot create tracer: %v", err)
	}
	appLogger.Info("Jaeger initialized")

	// Init Opentracing
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	appLogger.Info("Opentracing initialized")

	// Init MongoDB
	mongoDBConn, err := mongodb.NewMongoDBConn(ctx, cfg)
	if err != nil {
		appLogger.Fatal("cannot initialize MongoDB connection", err)
	}
	defer func() {
		if err := mongoDBConn.Disconnect(ctx); err != nil {
			appLogger.Fatal("cannot disconnect from MongoDB", err)
		}
	}()
	appLogger.Infof("MongoDB connected: %v", mongoDBConn.NumberSessionsInProgress())

	// Init Kafka
	conn, err := kafka.NewKafkaConnection(cfg)
	if err != nil {
		appLogger.Fatal("cannot create Kafka connection", err)
	}
	defer conn.Close()

	// Get brokers
	brokers, err := conn.Brokers()
	if err != nil {
		appLogger.Fatal("cannot get brokers", err)
	}
	appLogger.Infof("Kafka connected: %v", brokers)

	// Init Redis
	redisClient := redis.NewRedisClient(cfg)
	defer redisClient.Close()
	appLogger.Info("Redis connected")

	// Init Server
	s := server.NewServer(appLogger, cfg, tracer, mongoDBConn, redisClient)
	appLogger.Fatal(s.Start())
}
