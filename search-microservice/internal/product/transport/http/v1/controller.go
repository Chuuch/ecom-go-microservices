package v1

import (
	"net/http"

	"github.com/chuuch/search-microservice/config"
	"github.com/chuuch/search-microservice/internal/product/domain"
	httpErrors "github.com/chuuch/search-microservice/pkg/http_errors"
	"github.com/chuuch/search-microservice/pkg/logger"
	"github.com/chuuch/search-microservice/pkg/utils"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/opentracing/opentracing-go"
)

type productController struct {
	log            logger.Logger
	cfg            *config.Config
	productUsecase domain.ProductUsecase
	group          *echo.Group
	validate       *validator.Validate
}

func NewProductController(
	log logger.Logger,
	cfg *config.Config,
	productUsecase domain.ProductUsecase,
	group *echo.Group,
	validate *validator.Validate) *productController {
	return &productController{
		log:            log,
		cfg:            cfg,
		productUsecase: productUsecase,
		group:          group,
		validate:       validate,
	}
}

func (h *productController) index() echo.HandlerFunc {
	return func(c *echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(c.Request().Context(), "productController.index")
		defer span.Finish()

		var product domain.Product
		if err := c.Bind(&product); err != nil {
			h.log.Errorf("failed to bind product: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorResponse)
		}

		product.ID = uuid.New().String()

		if err := h.productUsecase.Index(ctx, product); err != nil {
			h.log.Error("failed to index product: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorResponse)
		}

		h.log.Info("product indexed successfully %s", product.ID)
		return c.JSON(http.StatusCreated, product)
	}
}

func (h *productController) search() echo.HandlerFunc {
	return func(c *echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(c.Request().Context(), "productController.Search")
		defer span.Finish()

		searchTerm := c.QueryParam("term")
		pagination := utils.NewPaginationFromQueryParams(c.QueryParam("size"), c.QueryParam("page"))

		searchResult, err := h.productUsecase.Search(ctx, searchTerm, pagination)
		if err != nil {
			h.log.Error("failed to search products: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorResponse)
		}

		h.log.Info("search products successfully %v", searchResult)
		return c.JSON(http.StatusOK, searchResult)
	}
}

func (h *productController) indexAsync() echo.HandlerFunc {
	return func(c *echo.Context) error {
		span, ctx := opentracing.StartSpanFromContext(c.Request().Context(), "productController.indexAsync")
		defer span.Finish()

		var product domain.Product
		if err := c.Bind(&product); err != nil {
			h.log.Error("failed to bind product: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorResponse)
		}

		product.ID = uuid.New().String()

		if err := h.productUsecase.IndexAsync(ctx, product); err != nil {
			h.log.Error("failed to indexproduct asynchronously: %v", err)
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorResponse)
		}

		h.log.Info("product indexed asynchronously %s", product.ID)
		return c.JSON(http.StatusCreated, product)
	}
}
