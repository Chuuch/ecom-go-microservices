package grpc

import (
	"context"

	"github.com/chuuch/product-microservice/internal/product"
	"github.com/chuuch/product-microservice/pkg/logger"
	productService "github.com/chuuch/product-microservice/proto/product"
	"github.com/go-playground/validator/v10"
)

// ProductGRPCService gRPC service
type ProductGRPCService struct {
	productService.UnimplementedProductServiceServer
	productUC product.UseCase
	log       logger.Logger
	validate  *validator.Validate
}

// ProductGRPCService constructor
func NewProductGRPCService(productUC product.UseCase, log logger.Logger, validate *validator.Validate) *ProductGRPCService {
	return &ProductGRPCService{
		productUC: productUC,
		log:       log,
		validate:  validate,
	}
}

func (p *ProductGRPCService) CreateProduct(ctx context.Context, req *productService.CreateRequest) (*productService.CreateResponse, error) {
	panic("not implemented")
}

func (p *ProductGRPCService) UpdateProduct(ctx context.Context, req *productService.UpdateRequest) (*productService.UpdateResponse, error) {
	panic("not implemented")
}

func (p *ProductGRPCService) GetProductByID(ctx context.Context, req *productService.FindByIDRequest) (*productService.FindByIDResponse, error) {
	panic("not implemented")
}

func (p *ProductGRPCService) SearchProducts(ctx context.Context, req *productService.SearchRequest) (*productService.SearchResponse, error) {
	panic("not implemented")
}
