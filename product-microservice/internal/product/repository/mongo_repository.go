package repository

import (
	"context"
	"log"
	"time"

	"github.com/chuuch/product-microservice/internal/models"
	productErrors "github.com/chuuch/product-microservice/pkg/product_errors"
	"github.com/chuuch/product-microservice/pkg/utils"
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

func (p *productMongoRepo) SearchProducts(ctx context.Context, query string, pagination *utils.Pagination) (*models.ProductsList, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "productMongoRepo.SearchProducts")
	defer span.Finish()

	collection := p.mongoDB.Database(productsDB).Collection(productsCollection)

	f := bson.D{
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "name", Value: primitive.Regex{Pattern: query, Options: "i"}}},
			bson.D{{Key: "description", Value: primitive.Regex{Pattern: query, Options: "i"}}},
		}},
	}

	count, err := collection.CountDocuments(ctx, f)
	if err != nil {
		return nil, errors.Wrap(err, "CountDocuments failed")
	}

	if count == 0 {
		return &models.ProductsList{
			TotalCount: 0,
			TotalPages: 0,
			Page:       0,
			Size:       0,
			HasMore:    false,
			Products:   make([]*models.Product, 0),
		}, nil
	}

	limit := int64(pagination.GetLimit())
	skip := int64(pagination.GetOffset())
	cursor, err := collection.Find(ctx, f, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Find failed")
	}
	defer cursor.Close(ctx)

	products := make([]*models.Product, 0, pagination.GetSize())
	for cursor.Next(ctx) {
		var prod models.Product
		if err := cursor.Decode(&prod); err != nil {
			return nil, errors.Wrap(err, "Decode failed")
		}
		products = append(products, &prod)
	}

	if err := cursor.Err(); err != nil {
		return nil, errors.Wrap(err, "Find failed")
	}

	log.Printf("SEARCH PRODUCTS: %+v", products)

	return &models.ProductsList{
		TotalCount: count,
		TotalPages: int64(pagination.GetTotalPages(int(count))),
		Page:       int64(pagination.GetPage()),
		Size:       int64(pagination.GetSize()),
		HasMore:    pagination.GetHasMore(int(count)),
		Products:   products,
	}, nil
}
