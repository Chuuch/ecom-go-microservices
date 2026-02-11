package product

import (
	"context"

	"github.com/chuuch/product-microservice/internal/models"
)

// Product repository interface
type MongoRepository interface {
	CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error)
	UpdateProduct(ctx context.Context, product *models.Product) (*models.Product, error)
	GetProductByID(ctx context.Context, productID string) (*models.Product, error)
	SearchProducts(ctx context.Context, query string, page, size int64) ([]*models.Product, error)
}
