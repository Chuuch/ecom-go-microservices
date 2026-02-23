package main

import (
	"log"

	"github.com/chuuch/search-microservice/config"
	"github.com/chuuch/search-microservice/internal/app"
	"github.com/chuuch/search-microservice/pkg/jaeger"
	"github.com/chuuch/search-microservice/pkg/logger"
	"github.com/opentracing/opentracing-go"
)

func main() {
	log.Println("Starting search microservice")

	cfgFile, err := config.LoadConfig("./config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfg, err := config.ParseConfig(cfgFile)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	apiLogger := logger.NewApiLogger(cfg)
	apiLogger.InitLogger()
	apiLogger.Info("config", "cfg", cfg)
	apiLogger.Info("Logger initialized")

	tracer, closer, err := jaeger.InitJaeger(cfg)
	if err != nil {
		apiLogger.Fatalf("Failed to initialize Jaeger: %v", err)
	}
	defer closer.Close()

	opentracing.SetGlobalTracer(tracer)

	apiLogger.Info("Jaeger initialized")
	apiLogger.Fatal(app.NewApp(apiLogger, cfg).Run())
}
