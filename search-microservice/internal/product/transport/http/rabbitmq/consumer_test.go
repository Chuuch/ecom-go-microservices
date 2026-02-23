package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/chuuch/search-microservice/config"
	"github.com/chuuch/search-microservice/internal/product/domain"
	"github.com/chuuch/search-microservice/pkg/logger"
	"github.com/chuuch/search-microservice/pkg/rabbitmq"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

func TestConsumer_ConsumeIndexDeliveries(t *testing.T) {
	t.Parallel()

	cfgFile, err := config.LoadConfig("../../../../../config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cfg, err := config.ParseConfig(cfgFile)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	apiLogger := logger.NewApiLogger(cfg)
	apiLogger.InitLogger()

	amqpPublisher, err := rabbitmq.NewPublisher(cfg, apiLogger)
	if err != nil {
		require.NoError(t, err)
	}

	defer amqpPublisher.Close()

	for i := 0; i < 500; i++ {
		time.Sleep(50 * time.Microsecond)

		product := domain.Product{
			ID:           uuid.New().String(),
			Title:        gofakeit.Name(),
			Description:  gofakeit.Sentence(10),
			ImageURL:     gofakeit.URL(),
			CountInStock: gofakeit.Number(1, 100),
			Shop:         gofakeit.Company(),
			CreatedAt:    time.Now().UTC(),
		}

		dataBytes, err := json.Marshal(&product)
		if err != nil {
			require.NoError(t, err)
		}

		if err := amqpPublisher.Publish(
			context.Background(),
			cfg.RabbitMQ.ExchangeName,
			cfg.RabbitMQ.BindingKey,
			amqp091.Publishing{
				Body:        dataBytes,
				ContentType: echo.MIMEApplicationJSON,
				Headers:     map[string]interface{}{"trace-id": uuid.New().String()},
			},
		); err != nil {
			require.NoError(t, err)
		}

		apiLogger.Info("product published successfully %s", product.ID)
	}
}
