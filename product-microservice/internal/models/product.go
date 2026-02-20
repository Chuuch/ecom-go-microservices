package models

import (
	"time"

	productService "github.com/chuuch/product-microservice/proto/product"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Product model
type Product struct {
	ProductID   primitive.ObjectID `json:"product_id" bson:"_id,omitempty"`
	CategoryID  primitive.ObjectID `json:"category_id" bson:"category_id,omitempty"`
	Name        string             `json:"name" bson:"name,omitempty" validate:"required,min=3,max=100"`
	Description string             `json:"description"`
	Price       float64            `json:"price" bson:"price,omitempty" validate:"required"`
	ImageURL    *string            `json:"image_url" bson:"image_url,omitempty"`
	Photos      []string           `json:"photos" bson:"photos,omitempty"`
	Quantity    int64              `json:"quantity" bson:"quantity,omitempty" validate:"required"`
	Rating      int64              `json:"rating" bson:"rating,omitempty" validate:"required,min=0,max=10"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at,omitempty"`
}

// Get Image url helper
func (p *Product) GetImageURL() string {
	var imageURL string
	if p.ImageURL != nil {
		imageURL = *p.ImageURL
	}
	return imageURL
}

// ToProto Convert product to proto
func (p *Product) ToProto() *productService.Product {
	return &productService.Product{
		ProductId:   p.ProductID.String(),
		CategoryId:  p.CategoryID.String(),
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		ImageUrl:    p.GetImageURL(),
		Photos:      p.Photos,
		Quantity:    p.Quantity,
		Rating:      int64(p.Rating),
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

// Product from proto to model
func ProductFromProto(product *productService.Product) (*Product, error) {
	productID, err := primitive.ObjectIDFromHex(product.GetProductId())
	if err != nil {
		return nil, errors.Wrap(err, "primitive.ObjectIDFromHex")
	}

	catID, err := primitive.ObjectIDFromHex(product.GetCategoryId())
	if err != nil {
		return nil, errors.Wrap(err, "primitive.ObjectIDFromHex failed")
	}

	return &Product{
		ProductID:   productID,
		CategoryID:  catID,
		Name:        product.GetName(),
		Description: product.GetDescription(),
		Price:       product.GetPrice(),
		ImageURL:    &product.ImageUrl,
		Photos:      product.GetPhotos(),
		Quantity:    product.GetQuantity(),
		Rating:      int64(product.GetRating()),
		CreatedAt:   product.GetCreatedAt().AsTime(),
		UpdatedAt:   product.GetUpdatedAt().AsTime(),
	}, nil
}

// ProductsList All products response with pagination
type ProductsList struct {
	TotalCount int64      `json:"total_count"`
	TotalPages int64      `json:"total_pages"`
	Page       int64      `json:"page"`
	Size       int64      `json:"size"`
	HasMore    bool       `json:"has_more"`
	Products   []*Product `json:"products"`
}

// ToProtoList convert products list to proto
func (p *ProductsList) ToProtoList() []*productService.Product {
	// Pre-allocate memory for efficiency of length 0 with capacity for all items.
	productList := make([]*productService.Product, 0, len(p.Products))
	for _, product := range p.Products {
		productList = append(productList, product.ToProto())
	}
	return productList
}
