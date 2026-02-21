package esclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type ElasticIndex struct {
	Path  string `mapstructure:"path" validated:"required"`
	Name  string `mapstructure:"name" validated:"required"`
	Alias string `mapstructure:"alias" validated:"required"`
}

func (e *ElasticIndex) String() string {
	return fmt.Sprintf("Path: %s, Name: %s, Alias: %s", e.Path, e.Name, e.Alias)
}

func Info(ctx context.Context, esClient *elasticsearch.Client) (*esapi.Response, error) {
	response, err := esClient.Info(esClient.Info.WithContext(ctx), esClient.Info.WithHuman())
	if err != nil {
		return nil, err
	}
	if response.IsError() {
		return nil, errors.New(response.String())
	}
	return response, nil
}

func CreateIndex(ctx context.Context, esClient *elasticsearch.Client, name string, data []byte) (*esapi.Response, error) {
	response, err := esClient.Indices.Create(
		name,
		esClient.Indices.Create.WithContext(ctx),
		esClient.Indices.Create.WithBody(bytes.NewReader(data)),
		esClient.Indices.Create.WithPretty(),
		esClient.Indices.Create.WithHuman(),
		esClient.Indices.Create.WithTimeout(3*time.Second),
	)

	if err != nil {
		return nil, err
	}

	if response.IsError() {
		return nil, errors.New(response.String())
	}

	return response, nil
}

func CreateAlias(ctx context.Context, esClient *elasticsearch.Client, indexes []string, name string, data []byte) (*esapi.Response, error) {
	response, err := esClient.Indices.PutAlias(
		indexes,
		name,
		esClient.Indices.PutAlias.WithContext(ctx),
		esClient.Indices.PutAlias.WithBody(bytes.NewReader(data)),
		esClient.Indices.PutAlias.WithPretty(),
		esClient.Indices.PutAlias.WithHuman(),
		esClient.Indices.PutAlias.WithTimeout(3*time.Second),
	)

	if err != nil {
		return nil, err
	}

	if response.IsError() {
		return nil, errors.New(response.String())
	}

	return response, nil
}

func Exists(ctx context.Context, esClient *elasticsearch.Client, indexes []string) (*esapi.Response, error) {
	response, err := esClient.Indices.Exists(
		indexes,
		esClient.Indices.Exists.WithContext(ctx),
		esClient.Indices.Exists.WithPretty(),
		esClient.Indices.Exists.WithHuman(),
	)

	if err != nil {
		return nil, err
	}

	if response.IsError() && response.StatusCode != 404 {
		return nil, errors.New(response.String())
	}

	return response, nil
}
