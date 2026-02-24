package utils

import (
	"fmt"
	"math"
	"strconv"
)

const (
	defaultSize = 2
	defaultPage = 1
)

type PaginationResponse struct {
	TotalCount int64 `json:"totalCount"`
	TotalPages int64 `json:"totalPages"`
	Page       int64 `json:"Page"`
	Size       int64 `json:"Size"`
	HasMore    bool  `json:"HasMore"`
}

func (p *PaginationResponse) String() string {
	return fmt.Sprintf("TotalCount: %d, TotalPages: %d, Size: %d, HasMore: %t", p.TotalCount, p.TotalPages, p.Page, p.Size, p.HasMore)
}

type Pagination struct {
	Size    int64  `json:"size,omitempty"`
	Page    int64  `json:"page,omitempty"`
	OrderBy string `json:"orderby,omitempty"`
}

func NewPaginationResponse(totalCount int64, pq *Pagination) *PaginationResponse {
	return &PaginationResponse{
		TotalCount: totalCount,
		TotalPages: pq.GetTotalPages(totalCount),
		Page:       pq.GetPage(),
		Size:       pq.GetSize(),
		HasMore:    pq.GetHasMore(totalCount),
	}
}

func NewPaginationQuery(size int64, page int64) *Pagination {
	if size == 0 {
		return &Pagination{
			Size: defaultSize,
			Page: defaultPage,
		}
	}
	return &Pagination{
		Size: size,
		Page: page,
	}
}

func NewPaginationFromQueryParams(size string, page string) *Pagination {
	p := &Pagination{
		Size: defaultSize,
		Page: 1,
	}

	if sizeNum, err := strconv.Atoi(size); err == nil && sizeNum != 0 {
		p.Size = int64(sizeNum)
	}

	if pageNum, err := strconv.Atoi(page); err == nil && pageNum != 0 {
		p.Page = int64(pageNum)
	}

	return p
}

func (p *Pagination) SetSize(sizeQuery string) error {
	if sizeQuery == "" {
		p.Size = defaultSize
		return nil
	}

	sizeNum, err := strconv.Atoi(sizeQuery)
	if err != nil {
		return err
	}

	p.Size = int64(sizeNum)
	return nil
}

func (p *Pagination) SetPage(pageQuery string) error {
	if pageQuery == "" {
		p.Page = defaultPage
		return nil
	}

	pageNum, err := strconv.Atoi(pageQuery)
	if err != nil {
		return err
	}

	p.Page = int64(pageNum)
	return nil
}

func (p *Pagination) SetOrderBy(orderByQuery string) error {
	if orderByQuery == "" {
		p.OrderBy = ""
		return nil
	}

	p.OrderBy = orderByQuery
	return nil
}

func (p *Pagination) GetOffset() int64 {
	if p.Page == 0 {
		return 0
	}
	return (p.Page - 1) * p.Size
}

func (p *Pagination) GetLimit() int64 {
	return p.Size
}

func (p *Pagination) GetOrderBy() string {

	return p.OrderBy
}

func (p *Pagination) GetPage() int64 {
	return p.Page
}

func (p *Pagination) GetSize() int64 {
	return p.Size
}

func (p *Pagination) GetQueryString() string {
	return fmt.Sprintf("page=%d&size=%d&orderBy=%s", p.GetPage(), p.GetSize(), p.GetOrderBy())
}

func (p *Pagination) GetTotalPages(totalCount int64) int64 {
	d := float64(totalCount) / float64(p.GetSize())
	return int64(math.Ceil(d))
}

func (p *Pagination) GetHasMore(totalCount int64) bool {
	return p.GetPage() < totalCount/p.GetSize()
}
