package config

import (
	"errors"
	"log"
	"net/http"

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

	return &c, nil
}

type Config struct {
	Server  ServerConfig
	Logger  LoggerConfig
	Jaeger  JaegerConfig
	Elastic ElasticConfig
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
}

type ElasticConfig struct {
	Addresses     []string
	Username      string
	Password      string
	APIKey        string
	Header        http.Header
	EnableLogging bool
}
