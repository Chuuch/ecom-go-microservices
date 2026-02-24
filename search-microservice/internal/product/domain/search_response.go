package domain

import "github.com/chuuch/search-microservice/pkg/utils"

type ProductSearchResponse struct {
	List               []*Product                `json:"list"`
	PaginationResponse *utils.PaginationResponse `json:"paginationResponse"`
}
