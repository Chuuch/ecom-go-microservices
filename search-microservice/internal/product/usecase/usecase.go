package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/chuuch/search-microservice/config"
	"github.com/chuuch/search-microservice/internal/product/domain"
	"github.com/chuuch/search-microservice/pkg/logger"
	"github.com/chuuch/search-microservice/pkg/rabbitmq"
	"github.com/chuuch/search-microservice/pkg/utils"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
)

type productUsecase struct {
	log               logger.Logger
	cfg               *config.Config
	elasticRepository domain.ElasticRepository
	amqpPublisher     rabbitmq.AmqpPublisher
}

func NewProductUsecase(
	log logger.Logger,
	cfg *config.Config,
	elasticRepository domain.ElasticRepository,
	amqpPublisher rabbitmq.AmqpPublisher,
) *productUsecase {
	return &productUsecase{
		log:               log,
		cfg:               cfg,
		elasticRepository: elasticRepository,
		amqpPublisher:     amqpPublisher,
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

func (u *productUsecase) IndexAsync(ctx context.Context, product domain.Product) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUsecase.IndexAsync")
	defer span.Finish()

	dataBytes, err := json.Marshal(&product)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	return u.amqpPublisher.Publish(
		ctx,
		u.cfg.RabbitMQ.ExchangeName,
		u.cfg.RabbitMQ.BindingKey,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        dataBytes,
			Timestamp:   time.Now().UTC(),
		},
	)
}
