package product

import (
	"context"

	"github.com/chuuch/product-microservice/internal/models"
	"github.com/chuuch/product-microservice/pkg/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UseCase product
type UseCase interface {
	CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error)
	UpdateProduct(ctx context.Context, product *models.Product) (*models.Product, error)
	GetProductByID(ctx context.Context, productID primitive.ObjectID) (*models.Product, error)
	SearchProducts(ctx context.Context, query string, pagination *utils.Pagination) (*models.ProductsList, error)
}
