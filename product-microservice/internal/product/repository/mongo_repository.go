package repository

import (
	"context"

	"github.com/chuuch/product-microservice/internal/models"
)

// ProductMongoRepo

type productMongoRepo struct {
}

// ProductMongo Constructor
func NewProductMongoRepository() *productMongoRepo {
	return &productMongoRepo{}
}

// NewProductMongoRepository
func (p *productMongoRepo) CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	panic("not implemented")
}

func (p *productMongoRepo) UpdateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	panic("not implemented")
}

func (p *productMongoRepo) GetProductByID(ctx context.Context, productID string) (*models.Product, error) {
	panic("not implemented")
}

func (p *productMongoRepo) SearchProducts(ctx context.Context, query string, page, size int64) ([]*models.Product, error) {
	panic("not implemented")
}
