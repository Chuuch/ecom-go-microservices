package domain

import (
	"context"

	"github.com/chuuch/search-microservice/pkg/utils"
)

type ProductUsecase interface {
	Index(ctx context.Context, product Product) error
	Search(ctx context.Context, term string, pagination *utils.Pagination) (*ProductSearchResponse, error)
	// IndexAsync(ctx context.Context, product Product) error
}
