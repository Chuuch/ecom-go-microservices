package config

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/chuuch/search-microservice/pkg/esclient"
	"github.com/chuuch/search-microservice/pkg/probes"
	"github.com/spf13/viper"
)

// Load config file from given path
func LoadConfig(filename string) (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigFile(filename)
	v.AddConfigPath(".")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, errors.New("config file not found")
		}
		return nil, err
	}

	// Bind env vars so K8s/Docker env override file values
	_ = v.BindEnv("RabbitMQ.URI", "RABBITMQ_URI")
	_ = v.BindEnv("Jaeger.Host", "JAEGER_HOST_PORT")
	return v, nil
}

// Parse config
func ParseConfig(v *viper.Viper) (*Config, error) {
	var c Config
	err := v.Unmarshal(&c)
	if err != nil {
		log.Printf("unable to decode into struct, %v", err)
		return nil, err
	}

	// ELASTIC_URL is a single URL; override Elastic.Addresses when set
	if url := os.Getenv("ELASTIC_URL"); url != "" {
		c.Elastic.Addresses = []string{strings.TrimSpace(url)}
	}
	// JAEGER_HOST_PORT may be "host:port"; ensure full collector URL for jaeger-client-go
	if host := os.Getenv("JAEGER_HOST_PORT"); host != "" {
		host = strings.TrimSpace(host)
		if !strings.HasPrefix(host, "http") {
			host = "http://" + host
		}
		if !strings.Contains(host, "/api/traces") {
			host = strings.TrimSuffix(host, "/") + "/api/traces"
		}
		c.Jaeger.Host = host
	}
	// RABBITMQ_URI is bound above; viper may not override nested on Unmarshal, so apply explicitly
	if uri := os.Getenv("RABBITMQ_URI"); uri != "" {
		c.RabbitMQ.URI = strings.TrimSpace(uri)
	}
	return &c, nil
}

type Config struct {
	Server         ServerConfig
	Logger         LoggerConfig
	Jaeger         JaegerConfig
	Elastic        ElasticConfig
	ElasticMapping ElasticMappingConfig
	Http           HTTPConfig
	RabbitMQ       RabbitMQConfig
	Probes         probes.Probes
}

type ServerConfig struct {
	AppVersion  string
	Port        string
	Development bool
}

type LoggerConfig struct {
	Level         string
	Development   bool
	DisableCaller bool
	Encoding      string
}

type JaegerConfig struct {
	Host        string
	ServiceName string
	LogSpans    bool
	Enable      bool
}

type ElasticConfig struct {
	Addresses     []string
	Username      string
	Password      string
	APIKey        string
	Header        http.Header
	EnableLogging bool
}

type ElasticMappingConfig struct {
	Path          string
	Name          string
	Alias         string
	ProductsIndex esclient.ElasticIndex
}

type HTTPConfig struct {
	Port               string
	DebugErrorResponse bool
	Development        bool
	IgnoredURIs        []string
	ProductsPath       string
}

type RabbitMQConfig struct {
	URI          string
	ExchangeName string
	ExchangeKind string
	QueueName    string
	BindingKey   string
	Concurrency  int
	Consumer     string
	BulkIndexer  BulkIndexerConfig
}

type BulkIndexerConfig struct {
	NumWorkers           int `mapstructure:"numWorkers" validated:"required"`
	FlushBytes           int `mapstructure:"flushBytes" validated:"required"`
	FlushIntervalSeconds int `mapstructure:"flushIntervalSeconds" validated:"required"`
	TimeoutMilliseconds  int `mapstructure:"timeoutMilliseconds" validated:"required"`
}
