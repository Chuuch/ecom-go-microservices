package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/chuuch/search-microservice/config"
	"github.com/chuuch/search-microservice/internal/product/domain"
	"github.com/chuuch/search-microservice/pkg/esclient"
	"github.com/chuuch/search-microservice/pkg/logger"
	misstypemanager "github.com/chuuch/search-microservice/pkg/misstype_manager"
	"github.com/chuuch/search-microservice/pkg/utils"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

type ElasticRepository struct {
	log             logger.Logger
	cfg             *config.Config
	esClient        *elasticsearch.Client
	misstypeManager misstypemanager.MisstypeManager
}

func NewElasticRepository(log logger.Logger, cfg *config.Config, esClient *elasticsearch.Client, misstypeManager misstypemanager.MisstypeManager) *ElasticRepository {
	return &ElasticRepository{
		log:             log,
		cfg:             cfg,
		esClient:        esClient,
		misstypeManager: misstypeManager,
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

func (e *ElasticRepository) Search(
	ctx context.Context,
	term string,
	pagination *utils.Pagination,
) (*domain.ProductSearchResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ElasticRepository.Search")
	defer span.Finish()
	span.LogFields(log.String("term", term), log.Object("pagination", pagination))

	shouldQuery := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"should": []map[string]any{
					{
						"multi_match": map[string]any{
							"query":  term,
							"fields": []string{"title", "description"},
						},
					},
					{
						"multi_match": map[string]any{
							"query":  e.misstypeManager.GetMissTypeWord(term),
							"fields": []string{"title", "description"},
						},
					},
				},
			},
		},
	}

	dataBytes, err := json.Marshal(&shouldQuery)

	e.log.Info("search query: %s", shouldQuery)
	e.log.Info("JSON query: %s", string(dataBytes))

	response, err := e.esClient.Search(
		e.esClient.Search.WithContext(ctx),
		e.esClient.Search.WithIndex(e.cfg.ElasticMapping.ProductsIndex.Name),
		e.esClient.Search.WithBody(bytes.NewReader(dataBytes)),
		e.esClient.Search.WithPretty(),
		e.esClient.Search.WithHuman(),
		e.esClient.Search.WithTimeout(5*time.Second),
		e.esClient.Search.WithSize(int(pagination.GetSize())),
		e.esClient.Search.WithFrom(int(pagination.GetOffset())),
	)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	if response.IsError() {
		return nil, errors.Wrap(errors.New(response.String()), "failed to search")
	}

	hits := esclient.ESHits[*domain.Product]{}
	err = json.NewDecoder(response.Body).Decode(&hits)
	if err != nil {
		return nil, err
	}
	e.log.Info("search response: %s", response.String())

	responseList := make([]*domain.Product, len(hits.Hits.Hits))
	for i, source := range hits.Hits.Hits {
		responseList[i] = source.Source
	}

	e.log.Info("search response list: %v", responseList)
	return &domain.ProductSearchResponse{
		List:               responseList,
		PaginationResponse: utils.NewPaginationResponse(hits.Hits.Total.Value, pagination),
	}, nil
}
