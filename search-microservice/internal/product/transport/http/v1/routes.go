package v1

func (h *productController) MapRoutes() {
	h.group.POST("", h.index())
	h.group.GET("/search", h.search())
	h.group.POST("/async", h.indexAsync())
}
