package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/chuuch/search-microservice/config"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/sync/errgroup"
)

type RabbitMQ struct {
	URI string `mapstructure:"uri" validated:"required"`
}

func NewRabbitMQ(cfg *config.Config) (*amqp091.Connection, error) {
	conn, err := amqp091.Dial(cfg.RabbitMQ.URI)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type ExchangeAndQueueBinding struct {
	ExchangeName string `mapstructure:"exchangeName" validated:"required"`
	ExchangeKind string `mapstructure:"exchangeKind" validated:"required"`
	QueueName    string `mapstructure:"queueName" validated:"required"`
	BindingKey   string `mapstructure:"bindingKey" validated:"required"`
	Concurrency  int    `mapstructure:"concurrency" validated:"required"`
	Consumer     string `mapstructure:"consumer" validated:"required"`
}

type ConsumeDeliveriesWorker interface {
	ConsumeDeliveries(ctx context.Context, deliveries <-chan amqp091.Delivery, workerID int) func() error
}

type DeliveriesConsumer func(ctx context.Context, deliveries <-chan amqp091.Delivery, workerID int) func() error

func NewRabbitMQConnection(cfg *config.Config) (*amqp091.Connection, error) {
	conn, err := amqp091.Dial(cfg.RabbitMQ.URI)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func DeclareBinding(ctx context.Context, channel *amqp091.Channel, exchangeAndQueueBinding ExchangeAndQueueBinding) (amqp091.Queue, error) {
	if err := DeclareExchange(ctx, channel, exchangeAndQueueBinding.ExchangeName, exchangeAndQueueBinding.ExchangeKind); err != nil {
		return amqp091.Queue{}, err
	}

	queue, err := DeclareQueue(ctx, channel, exchangeAndQueueBinding.QueueName)
	if err != nil {
		return amqp091.Queue{}, err
	}

	if err := BindQueue(ctx, channel, queue.Name, exchangeAndQueueBinding.ExchangeName, exchangeAndQueueBinding.BindingKey); err != nil {
		return amqp091.Queue{}, err
	}
	return queue, nil
}

func DeclareExchange(ctx context.Context, channel *amqp091.Channel, name, kind string) error {
	return channel.ExchangeDeclare(
		name,
		kind,
		true,
		false,
		false,
		false,
		nil,
	)
}

func DeclareQueue(ctx context.Context, channel *amqp091.Channel, name string) (amqp091.Queue, error) {
	return channel.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
}

func BindQueue(ctx context.Context, channel *amqp091.Channel, queueName, exchangeName, bindingKey string) error {
	return channel.QueueBind(
		queueName,
		bindingKey,
		exchangeName,
		false,
		nil,
	)
}

func ConsumeQueue(ctx context.Context, channel *amqp091.Channel, concurrency int, queue string, consumer string, worker DeliveriesConsumer) error {
	deliveries, err := channel.Consume(
		queue,
		consumer,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(ctx)
	for i := 0; i < concurrency; i++ {
		eg.Go(worker(ctx, deliveries, i))
	}
	return eg.Wait()
}

func prcessDeliveries(ctx context.Context, deliveries <-chan amqp091.Delivery) func() error {
	return func() error {
		for delivery := range deliveries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			log.Printf("received message: %s", delivery.Body)
			if err := delivery.Ack(true); err != nil {
				return err
			}
		}
		return nil
	}
}

func Publish(ctx context.Context, channel *amqp091.Channel, exchange, key string, data any, headers map[string]any) error {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	amqpHeaders := amqp091.Table{}
	if headers != nil {
		for k, v := range headers {
			amqpHeaders[k] = v
		}
	}

	return channel.PublishWithContext(
		ctx,
		exchange,
		key,
		false,
		true,
		amqp091.Publishing{
			Headers:       amqpHeaders,
			ContentType:   echo.MIMEApplicationJSON,
			DeliveryMode:  2,
			Priority:      9,
			CorrelationId: uuid.New().String(),
			MessageId:     uuid.New().String(),
			Timestamp:     time.Now().UTC(),
			Body:          dataBytes,
		},
	)
}
