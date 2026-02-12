package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/chuuch/product-microservice/config"
	"github.com/chuuch/product-microservice/pkg/jaeger"
	"github.com/chuuch/product-microservice/pkg/logger"
	"github.com/chuuch/product-microservice/pkg/mongodb"
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
	appLogger.Info("MongoDB connected: %v", mongoDBConn.NumberSessionsInProgress())

	// Init HTTP server
	http.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("REQUEST: %v", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"message": "Hello world!"}); err != nil {
			log.Printf("ERROR: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	log.Printf("Server is listening on port: %v", cfg.Server.Port)
	log.Fatal(http.ListenAndServe(cfg.Server.Port, nil))
}
