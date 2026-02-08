package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/chuuch/product-microservice/config"
	"github.com/chuuch/product-microservice/pkg/jaeger"
	"github.com/chuuch/product-microservice/pkg/logger"
	"github.com/opentracing/opentracing-go"
)

func main() {
	log.Printf("Starting product microservice")

	c, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("ParseConfig: %v", err)
	}

	appLogger := logger.NewApiLogger(c)
	appLogger.InitLogger()
	appLogger.Info("Logger initialized")
	appLogger.Infof(
		"AppVersion: %s, LogLevel: %s, DevelopmentMode: %s",
		c.AppVersion,
		c.Logger.Level,
		c.Server.Development,
	)
	appLogger.Infof("Successfully parsed config: %v", c.AppVersion)

	tracer, closer, err := jaeger.InitJaeger(c)
	if err != nil {
		appLogger.Fatalf("cannot create tracer: %v", err)
	}
	appLogger.Info("Jaeger initialized")

	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	appLogger.Info("Opentracing initialized")

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
	log.Printf("Server is listening on port: %v", c.Server.Port)
	log.Fatal(http.ListenAndServe(c.Server.Port, nil))
}
