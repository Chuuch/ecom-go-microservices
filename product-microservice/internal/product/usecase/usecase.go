package usecase

import (
	"context"

	"github.com/chuuch/product-microservice/internal/models"
	"github.com/chuuch/product-microservice/internal/product"
	"github.com/chuuch/product-microservice/pkg/logger"
	"github.com/opentracing/opentracing-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type productUC struct {
	productRepo product.MongoRepository
	log         logger.Logger
}

func NewProductUC(productRepo product.MongoRepository, log logger.Logger) *productUC {
	return &productUC{
		productRepo: productRepo,
		log:         log,
	}
}

func (u *productUC) CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productUC.CreateProduct")
	defer span.Finish()

	return u.productRepo.CreateProduct(ctx, product)
}

func (u *productUC) UpdateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	panic("not implemented")
}

func (u *productUC) GetProductByID(ctx context.Context, productID primitive.ObjectID) (*models.Product, error) {
	panic("not implemented")
}

func (u *productUC) SearchProducts(ctx context.Context, query string, page, size int64) ([]*models.Product, error) {
	panic("not implemented")
}
