package repository

import (
	"context"
	"log"
	"time"

	"github.com/chuuch/product-microservice/internal/models"
	productErrors "github.com/chuuch/product-microservice/pkg/product_errors"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
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
	span, ctx := opentracing.StartSpanFromContext(ctx, "productMongoRepo.CreateProduct")
	defer span.Finish()

	collection := p.mongoDB.Database(productsDB).Collection(productsCollection)

	product.CreatedAt = time.Now().UTC()
	product.UpdatedAt = time.Now().UTC()

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
	span, ctx := opentracing.StartSpanFromContext(ctx, "productMongoRepo.UpdateProduct")
	defer span.Finish()

	collection := p.mongoDB.Database(productsDB).Collection(productsCollection)

	upsert := true
	after := options.After
	opts := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	}

	product.UpdatedAt = time.Now().UTC()

	var prod models.Product
	if err := collection.FindOneAndUpdate(ctx, bson.M{"_id": product.ProductID}, bson.M{"$set": product}, &opts).Decode(&prod); err != nil {
		return nil, errors.Wrap(err, "FindOneAndUpdate failed")
	}

	log.Printf("UPDATE PRODUCT: %+v", prod)

	return &prod, nil
}

func (p *productMongoRepo) GetProductByID(ctx context.Context, productID primitive.ObjectID) (*models.Product, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productMongoRepo.GetProductByID")
	defer span.Finish()

	collection := p.mongoDB.Database(productsDB).Collection(productsCollection)

	var prod models.Product
	if err := collection.FindOne(ctx, bson.M{"_id": productID}).Decode(&prod); err != nil {
		return nil, errors.Wrap(err, "FindOne failed")
	}

	log.Printf("GET PRODUCT: %+v", prod)

	return &prod, nil
}

func (p *productMongoRepo) SearchProducts(ctx context.Context, query string, page, size int64) ([]*models.Product, error) {
	panic("not implemented")
}
