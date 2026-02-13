package grpc

import (
	"context"

	"github.com/chuuch/product-microservice/internal/models"
	"github.com/chuuch/product-microservice/internal/product"
	grpcerrors "github.com/chuuch/product-microservice/pkg/grpc_errors"
	"github.com/chuuch/product-microservice/pkg/logger"
	productService "github.com/chuuch/product-microservice/proto/product"
	"github.com/go-playground/validator/v10"
	"github.com/opentracing/opentracing-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	span, ctx := opentracing.StartSpanFromContext(ctx, "ProductGRPCService.CreateProduct")
	defer span.Finish()

	catID, err := primitive.ObjectIDFromHex(req.GetCategoryId())
	if err != nil {
		p.log.Errorf("primitive.ObjectIDFromHex: %v", err)
		return nil, grpcerrors.ErrorResponse(err, err.Error())
	}

	product := &models.Product{
		CategoryID:  catID,
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Price:       req.GetPrice(),
		ImageURL:    &req.ImageUrl,
		Photos:      req.GetPhotos(),
		Quantity:    req.GetQuantity(),
		Rating:      int64(req.GetRating()),
	}

	if err := p.validate.StructCtx(ctx, product); err != nil {
		p.log.Errorf("validate.StructCtx: %v", err)
		return nil, grpcerrors.ErrorResponse(err, err.Error())
	}

	created, err := p.productUC.CreateProduct(ctx, product)
	if err != nil {
		p.log.Errorf("productUC.CreateProduct: %v", err)
		return nil, grpcerrors.ErrorResponse(err, err.Error())
	}
	return &productService.CreateResponse{
		Product: created.ToProto(),
	}, nil
}

func (p *ProductGRPCService) UpdateProduct(ctx context.Context, req *productService.UpdateRequest) (*productService.UpdateResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ProductGRPCService.UpdateProduct")
	defer span.Finish()

	productID, err := primitive.ObjectIDFromHex(req.GetProductId())
	if err != nil {
		p.log.Errorf("primitive.ObjectIDFromHex: %v", err)
		return nil, grpcerrors.ErrorResponse(err, err.Error())
	}

	catID, err := primitive.ObjectIDFromHex(req.GetCategoryId())
	if err != nil {
		p.log.Errorf("primitve.ObjectIDFromHex: %v", err)
		return nil, grpcerrors.ErrorResponse(err, err.Error())
	}

	product := &models.Product{
		ProductID:   productID,
		CategoryID:  catID,
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Price:       req.GetPrice(),
		ImageURL:    &req.ImageUrl,
		Photos:      req.GetPhotos(),
		Quantity:    req.GetQuantity(),
		Rating:      int64(req.GetRating()),
	}

	updated, err := p.productUC.UpdateProduct(ctx, product)
	if err != nil {
		p.log.Errorf("productUC.UpdateProduct: %v", err)
		return nil, grpcerrors.ErrorResponse(err, err.Error())
	}

	return &productService.UpdateResponse{
		Product: updated.ToProto(),
	}, nil
}

func (p *ProductGRPCService) FindByID(ctx context.Context, req *productService.FindByIDRequest) (*productService.FindByIDResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ProductGRPCService.FindByID")
	defer span.Finish()

	productID, err := primitive.ObjectIDFromHex(req.GetProductId())
	if err != nil {
		p.log.Errorf("primitve.ObjectIDFromHex: %v", err)
		return nil, grpcerrors.ErrorResponse(err, err.Error())
	}

	product, err := p.productUC.GetProductByID(ctx, productID)
	if err != nil {
		p.log.Errorf("productUC.GetProductByID: %v", err)
		return nil, grpcerrors.ErrorResponse(err, err.Error())
	}

	return &productService.FindByIDResponse{Product: product.ToProto()}, nil
}

func (p *ProductGRPCService) SearchProduct(ctx context.Context, req *productService.SearchRequest) (*productService.SearchResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ProductGRPCService.SearchProduct")
	defer span.Finish()

	products, err := p.productUC.SearchProducts(ctx, req.GetQuery(), req.GetPage(), req.GetPage())
	if err != nil {
		p.log.Errorf("productUC.SearchProducts: %v", err)
		return nil, grpcerrors.ErrorResponse(err, err.Error())
	}

	p.log.Info("productUC.SearchProducts: %v", products)
	return nil, nil
}
