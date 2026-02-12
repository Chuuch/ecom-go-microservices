package repository

import (
	"context"
	"log"
	"time"

	"github.com/chuuch/product-microservice/internal/models"
	productErrors "github.com/chuuch/product-microservice/pkg/product_errors"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	productsDB         = "products"
	productsCollection = "products"
)

// ProductMongoRepo

type productMongoRepo struct {
	mongoDB *mongo.Client
}

// ProductMongo Constructor
func NewProductMongoRepository(mongoDB *mongo.Client) *productMongoRepo {
	return &productMongoRepo{
		mongoDB: mongoDB,
	}
}

// NewProductMongoRepository
func (p *productMongoRepo) CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	collection := p.mongoDB.Database(productsDB).Collection(productsCollection)

	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	result, err := collection.InsertOne(ctx, product, &options.InsertOneOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "InsertOne.Collection")
	}

	objectID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.Wrap(productErrors.ErrObjectIDTypeConversion, "InsertOne.Collection")
	}

	product.ProductID = objectID
	log.Printf("CREATE PRODUCT: %+v", product)

	return product, nil
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
