package elastic

import (
	"net/http"
	"os"

	"github.com/chuuch/search-microservice/config"
	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
)

type Elastic struct {
	Addresses     []string `mapstructure:"addresses" validated:"required"`
	Username      string   `mapstructure:"username"`
	Password      string   `mapstructure:"password"`
	APIKey        string   `mapstructure:"apiKey"`
	Header        http.Header
	EnableLogging bool `mapstructure:"enableLogging"`
}

func NewElasticSearch(cfg *config.Config) (*elasticsearch.Client, error) {
	config := elasticsearch.Config{
		Addresses: cfg.Elastic.Addresses,
		Username:  cfg.Elastic.Username,
		Password:  cfg.Elastic.Password,
		APIKey:    cfg.Elastic.APIKey,
		Header:    cfg.Elastic.Header,
	}

	if cfg.Elastic.EnableLogging {
		config.Logger = &elastictransport.ColorLogger{
			Output:             os.Stdout,
			EnableRequestBody:  true,
			EnableResponseBody: true,
		}
	}
	client, err := elasticsearch.NewClient(config)
	if err != nil {
		return nil, err
	}

	return client, nil
}
