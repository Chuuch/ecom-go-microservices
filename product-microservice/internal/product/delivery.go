package product

import "github.com/labstack/echo/v4"

// HttpDelivery
type HttpDelivery interface {
	CreateProduct() echo.HandlerFunc
	UpdateProduct() echo.HandlerFunc
	GetProductByID() echo.HandlerFunc
	SearchProducts() echo.HandlerFunc
}
