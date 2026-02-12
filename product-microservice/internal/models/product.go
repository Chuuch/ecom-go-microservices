package models

import (
	"time"

	productService "github.com/chuuch/product-microservice/proto/product"
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
	ImageURL    string             `json:"image_url" bson:"image_url,omitempty"`
	Photos      []string           `json:"photos" bson:"photos,omitempty"`
	Quantity    int64              `json:"quantity" bson:"quantity,omitempty" validate:"required"`
	Rating      int64              `json:"rating" bson:"rating,omitempty" validate:"required,min=0,max=10"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at,omitempty"`
}

// ToProto Convert product to proto
func (p *Product) ToProto() *productService.Product {
	return &productService.Product{
		ProductId:   p.ProductID.String(),
		CategoryId:  p.CategoryID.String(),
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		ImageUrl:    p.ImageURL,
		Photos:      p.Photos,
		Quantity:    p.Quantity,
		Rating:      int64(p.Rating),
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}
