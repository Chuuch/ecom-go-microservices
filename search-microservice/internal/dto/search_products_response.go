package dto

import (
	"github.com/chuuch/search-microservice/internal/product/domain"
	"github.com/chuuch/search-microservice/pkg/utils"
)

type SearchProductsResponse struct {
	SearchTerm string            `json:"searchTerm"`
	Pagination *utils.Pagination `json:"pagination"`
	Products   []*domain.Product `json:"products"`
}
