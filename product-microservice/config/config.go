package config

import (
	"log"

	"github.com/spf13/viper"
)

// App config
type Config struct {
	AppVersion string
	Server     ServerConfig
	Logger     LoggerConfig
}

// Server config
type ServerConfig struct {
	Port        string
	Development bool
}

// Logger config
type LoggerConfig struct {
	DisableCaller     bool
	DisableStackTrace bool
	Encoding          string
	Level             string
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
