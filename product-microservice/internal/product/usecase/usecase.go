package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/chuuch/product-microservice/internal/models"
	"github.com/chuuch/product-microservice/internal/product"

	productKafka "github.com/chuuch/product-microservice/internal/product/delivery/kafka"
	"github.com/chuuch/product-microservice/pkg/logger"
	"github.com/chuuch/product-microservice/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type productUC struct {
	productRepo      product.MongoRepository
	log              logger.Logger
	validate         *validator.Validate
	redisRepo        product.RedisRepository
	productsProducer productKafka.ProductsProducer
}

func NewProductUC(
	productRepo product.MongoRepository,
	log logger.Logger,
	validate *validator.Validate,
	redisRepo product.RedisRepository,
	productsProducer productKafka.ProductsProducer,
) *productUC {
	return &productUC{
		productRepo:      productRepo,
		log:              log,
		validate:         validate,
		redisRepo:        redisRepo,
		productsProducer: productsProducer,
	}
}

func (u *productUC) CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.CreateProduct")
	defer span.Finish()

	product, err := u.productRepo.CreateProduct(ctx, product)
	if err != nil {
		return nil, errors.Wrap(err, "productRepo.CreateProduct failed")
	}

	if err := u.redisRepo.SetProduct(ctx, product); err != nil {
		return nil, errors.Wrap(err, "redisRepo.SetProduct failed")
	}

	return product, nil
}

func (u *productUC) UpdateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.UpdateProduct")
	defer span.Finish()

	return u.productRepo.UpdateProduct(ctx, product)
}

func (u *productUC) GetProductByID(ctx context.Context, productID primitive.ObjectID) (*models.Product, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.GetProductByID")
	defer span.Finish()

	cached, err := u.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, errors.Wrap(err, "redisRepo.GetProductByID failed")
	}

	if cached != nil {
		return cached, nil
	}

	product, err := u.productRepo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, errors.Wrap(err, "productRepo.GetProductByID failed")
	}

	if err := u.redisRepo.SetProduct(ctx, product); err != nil {
		return nil, errors.Wrap(err, "redisRepo.SetProduct failed")
	}

	return product, nil
}

func (u *productUC) SearchProducts(ctx context.Context, query string, pagination *utils.Pagination) (*models.ProductsList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.SearchProducts")
	defer span.Finish()

	return u.productRepo.SearchProducts(ctx, query, pagination)
}

func (u *productUC) PublishCreate(ctx context.Context, product *models.Product) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.PublishCreate")
	defer span.Finish()

	productBytes, err := json.Marshal(&product)
	if err != nil {
		return errors.Wrap(err, "json.Marshal failed")
	}

	return u.productsProducer.PublishCreate(ctx, kafka.Message{
		Value: productBytes,
		Time:  time.Now().UTC(),
	})
}

func (u *productUC) PublishUpdate(ctx context.Context, product *models.Product) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.PublishUpdate")
	defer span.Finish()

	productBytes, err := json.Marshal(&product)
	if err != nil {
		return errors.Wrap(err, "json.Marshal failed")
	}

	return u.productsProducer.PublishUpdate(ctx, kafka.Message{
		Value: productBytes,
		Time:  time.Now().UTC(),
	})
}
