package usecase

import (
	"context"

	"github.com/chuuch/search-microservice/config"
	"github.com/chuuch/search-microservice/internal/product/domain"
	"github.com/chuuch/search-microservice/pkg/logger"
	"github.com/chuuch/search-microservice/pkg/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
)

type productUsecase struct {
	log               logger.Logger
	cfg               *config.Config
	elasticRepository domain.ElasticRepository
}

func NewProductUsecase(log logger.Logger, cfg *config.Config, elasticRepository domain.ElasticRepository) *productUsecase {
	return &productUsecase{
		log:               log,
		cfg:               cfg,
		elasticRepository: elasticRepository,
	}
}

func (u *productUsecase) Index(ctx context.Context, product domain.Product) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUsecase.Index")
	defer span.Finish()

	span.LogFields(log.Object("product", product))

	return u.elasticRepository.Index(ctx, product)
}

func (u *productUsecase) Search(ctx context.Context, term string, pagination *utils.Pagination) (*domain.ProductSearchResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUsecase.Search")
	defer span.Finish()

	span.LogFields(log.String("term", term), log.Object("pagination", pagination))
	return u.elasticRepository.Search(ctx, term, pagination)
}
