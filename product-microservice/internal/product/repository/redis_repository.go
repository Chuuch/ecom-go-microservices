package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chuuch/product-microservice/internal/models"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	prefix     = "products"
	expiration = time.Second * 3600
)

type productRedisRepo struct {
	prefix string
	redis  *redis.Client
}

// NewProductRedisRepository constructor
func NewProductRedisRepository(prefix string, redis *redis.Client) *productRedisRepo {
	return &productRedisRepo{
		prefix: prefix,
		redis:  redis,
	}
}

// SetProduct set product in redis
func (r *productRedisRepo) SetProduct(ctx context.Context, product *models.Product) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productRedisRepo.SetProduct")
	defer span.Finish()

	prodBytes, err := json.Marshal(product)
	if err != nil {
		return errors.Wrap(err, "json.Marshal failed")
	}

	return r.redis.Set(ctx, r.createKey(product.ProductID), string(prodBytes), expiration).Err()
}

// GetProductByID get product by id from redis
func (r *productRedisRepo) GetProductByID(ctx context.Context, productID primitive.ObjectID) (*models.Product, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productRedisRepo.GetProductByID")
	defer span.Finish()

	result, err := r.redis.Get(ctx, r.createKey(productID)).Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "redis.Get failed")
	}

	var res models.Product
	if err := json.Unmarshal(result, &res); err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal failed")
	}

	return &res, nil
}

func (r *productRedisRepo) createKey(productID primitive.ObjectID) string {
	return fmt.Sprintf("%s:%s", r.prefix, productID.String())
}
