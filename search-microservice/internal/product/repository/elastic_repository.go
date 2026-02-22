package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/chuuch/search-microservice/config"
	"github.com/chuuch/search-microservice/internal/product/domain"
	"github.com/chuuch/search-microservice/pkg/logger"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

type ElasticRepository struct {
	log      logger.Logger
	cfg      *config.Config
	esClient *elasticsearch.Client
}

func NewElasticRepository(log logger.Logger, cfg *config.Config, esClient *elasticsearch.Client) *ElasticRepository {
	return &ElasticRepository{
		log:      log,
		cfg:      cfg,
		esClient: esClient,
	}
}

func (e *ElasticRepository) Index(ctx context.Context, product domain.Product) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ElasticRepository.Index")
	defer span.Finish()

	span.LogFields(log.Object("product", product))
	dataBytes, err := json.Marshal(&product)
	if err != nil {
		return errors.Wrap(err, "json.Marshal failed")
	}

	response, err := e.esClient.Index(
		e.cfg.ElasticMapping.ProductsIndex.Name,
		bytes.NewReader(dataBytes),
		e.esClient.Index.WithContext(ctx),
		e.esClient.Index.WithPretty(),
		e.esClient.Index.WithHuman(),
		e.esClient.Index.WithTimeout(3*time.Second),
		e.esClient.Index.WithDocumentID(product.ID),
	)
	if err != nil {
		return errors.Wrap(err, "failed to index product")
	}

	defer response.Body.Close()
	if response.IsError() {
		return errors.New(response.String())
	}
	e.log.Info("product indexed successfully %s", response.String())
	return nil
}
