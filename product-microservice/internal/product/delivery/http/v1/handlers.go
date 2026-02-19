package v1

import (
	"net/http"
	"strconv"

	"github.com/chuuch/product-microservice/internal/middleware"
	"github.com/chuuch/product-microservice/internal/models"
	"github.com/chuuch/product-microservice/internal/product"
	httpErrors "github.com/chuuch/product-microservice/pkg/http_errors"
	"github.com/chuuch/product-microservice/pkg/logger"
	"github.com/chuuch/product-microservice/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type productHandlers struct {
	log       logger.Logger
	productUC product.UseCase
	validate  *validator.Validate
	group     *echo.Group
	mw        middleware.MiddlewareManager
}

func NewProductHandlers(log logger.Logger, productUC product.UseCase, validate *validator.Validate, group *echo.Group, mw middleware.MiddlewareManager) *productHandlers {
	return &productHandlers{
		log:       log,
		productUC: productUC,
		validate:  validate,
		group:     group,
		mw:        mw,
	}
}

func (h *productHandlers) CreateProduct() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(c.Request().Context(), "productHandlers.CreateProduct")
		defer span.Finish()

		createRequests.Inc()

		var prod models.Product
		if err := c.Bind(&prod); err != nil {
			h.log.Errorf("c.Bind: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		if err := h.validate.StructCtx(ctx, prod); err != nil {
			h.log.Errorf("validate.StructCtx: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		if err := h.productUC.PublishCreate(ctx, &prod); err != nil {
			h.log.Errorf("productUC.CreateProduct: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.NoContent(http.StatusCreated)
	}
}

func (h *productHandlers) UpdateProduct() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(c.Request().Context(), "productHandlers.UpdateRequest")
		defer span.Finish()

		updateRequests.Inc()

		var prod models.Product
		if err := c.Bind(&prod); err != nil {
			h.log.Errorf("c.Bind: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		productID, err := primitive.ObjectIDFromHex(c.Param("product_id"))
		if err != nil {
			h.log.Errorf("primitive.ObjectIDFromHex: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}
		prod.ProductID = productID

		if err := h.validate.StructCtx(ctx, prod); err != nil {
			h.log.Errorf("validate.StructCtx: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		if err := h.productUC.PublishUpdate(ctx, &prod); err != nil {
			h.log.Errorf("productUC.UpdateProduct: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.NoContent(http.StatusOK)
	}
}

func (h *productHandlers) GetProductByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(c.Request().Context(), "productHandlers.GetProductByID")
		defer span.Finish()

		getRequests.Inc()

		productID, err := primitive.ObjectIDFromHex(c.QueryParam("product_id"))
		if err != nil {
			h.log.Errorf("primitive.ObjectIDFromHex: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		product, err := h.productUC.GetProductByID(ctx, productID)
		if err != nil {
			h.log.Errorf("productUC.GetProductByID: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.JSON(http.StatusOK, product)
	}
}

func (h *productHandlers) SearchProducts() echo.HandlerFunc {
	return func(c echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(c.Request().Context(), "productHandlers.SearchProducts")
		defer span.Finish()

		searchRequests.Inc()

		page, err := strconv.Atoi(c.QueryParam("page"))
		if err != nil {
			h.log.Errorf("strconv.Atoi: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		size, err := strconv.Atoi(c.QueryParam("size"))
		if err != nil {
			h.log.Errorf("strconv.Atoi: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		pq := utils.NewPaginationQuery(size, page)

		result, err := h.productUC.SearchProducts(ctx, c.QueryParam("query"), pq)
		if err != nil {
			h.log.Errorf("productUC.SearchProducts: %v", err)
			errorRequests.Inc()
			return httpErrors.ErrorCtxResponse(c, err)
		}

		successRequests.Inc()
		return c.JSON(http.StatusBadRequest, result)
	}
}
