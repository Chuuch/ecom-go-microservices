package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// App config
type Config struct {
	AppVersion string
	Server     ServerConfig
	Logger     LoggerConfig
	Jaeger     JaegerConfig
	Metrics    MetricsConfig
	MongoDB    MongoDBConfig
	Kafka      KafkaConfig
	Http       HttpConfig
}

// Server config
type ServerConfig struct {
	Port              string
	Development       bool
	Timeout           time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	MaxConnectionIdle time.Duration
	MaxConnectionAge  time.Duration
}

// Logger config
type LoggerConfig struct {
	DisableCaller     bool
	DisableStackTrace bool
	Encoding          string
	Level             string
}

// Jaeger config
type JaegerConfig struct {
	Host        string
	ServiceName string
	LogSpans    bool
}

// Metrics config
type MetricsConfig struct {
	Port        string
	URL         string
	ServiceName string
}

// MongoDB config
type MongoDBConfig struct {
	URI      string
	User     string
	Password string
	DB       string
}

// Kafka config
type KafkaConfig struct {
	Brokers []string
}

// Http config
type HttpConfig struct {
	Port              string
	PprofPort         string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	Timeout           time.Duration
	CookieLifetime    time.Duration
	SessionCookieName string
}

// Load config file from given path
func exportConfig() error {
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.SetConfigName("config.yaml")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

// Parse config
func ParseConfig() (*Config, error) {
	if err := exportConfig(); err != nil {
		return nil, err
	}

	var c Config
	err := viper.Unmarshal(&c)
	if err != nil {
		log.Printf("unable to decode into struct, %v", err)
		return nil, err
	}

	return &c, nil
}
