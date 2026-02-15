package v1

// MapRoutes products routes
func (h *productHandlers) MapRoutes() {
	h.group.POST("", h.CreateProduct())
	h.group.PUT("/:product_id", h.UpdateProduct())
	h.group.GET("/:product_id", h.GetProductByID())
	h.group.GET("/search", h.SearchProducts())
}
