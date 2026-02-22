package domain

import (
	"context"

	"github.com/chuuch/search-microservice/pkg/utils"
)

type ElasticRepository interface {
	Index(ctx context.Context, product Product) error
	Search(ctx context.Context, term string, pagination *utils.Pagination) (*ProductSearchResponse, error)
}
