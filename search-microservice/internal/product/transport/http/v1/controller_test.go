package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/chuuch/search-microservice/internal/product/domain"
	"github.com/chuuch/search-microservice/pkg/http_client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIndexProduct(t *testing.T) {
	t.Parallel()

	client := http_client.NewHttpClient(true)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	product := domain.Product{
		ID:           uuid.New().String(),
		Title:        "Test Product",
		Description:  "Test Description",
		ImageURL:     "https://example.com/image.jpg",
		CountInStock: 10,
		Shop:         "Test Shop",
		CreatedAt:    time.Now().UTC(),
	}

	t.Logf("indexing product %s", product.ID)

	response, err := client.R().
		SetBody(product).
		SetContext(ctx).
		Post("http://localhost:8000/v1/products")

	require.NoError(t, err)
	require.NotNil(t, response)
	require.False(t, response.IsError())
	require.True(t, response.IsSuccess())
	require.Equal(t, response.StatusCode(), http.StatusCreated)

	var productResponse domain.Product
	err = json.Unmarshal(response.Body(), &productResponse)
	require.NoError(t, err)
	require.NotNil(t, productResponse)
	require.NotEmpty(t, productResponse.ID)

	t.Logf("product indexed successfully %s", productResponse.ID)
}
