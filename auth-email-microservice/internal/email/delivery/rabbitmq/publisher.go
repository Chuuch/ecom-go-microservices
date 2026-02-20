package rabbitmq

import (
	"time"

	"github.com/Chuuch/ecom-microservices/config"
	"github.com/Chuuch/ecom-microservices/pkg/logger"
	"github.com/Chuuch/ecom-microservices/pkg/rabbitmq"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/streadway/amqp"
)

var (
	publishedMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "email_published_rabbitmq_messages_total",
		Help: "The total number of published RabbitMQ messages",
	})
)

// Emails rabbitmq publisher
type EmailsPublisher struct {
	amqpChan *amqp.Channel
	cfg      *config.Config
	logger   logger.Logger
}

// Emails rabbitmq publisher constructor
func NewEmailsPublisher(cfg *config.Config, logger logger.Logger) (*EmailsPublisher, error) {
	mqConn, err := rabbitmq.NewRabbitMQConnection(cfg)
	if err != nil {
		return nil, err
	}
	amqpChan, err := mqConn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "p.mqConn.Channel")
	}
	return &EmailsPublisher{
		amqpChan: amqpChan,
		cfg:      cfg,
		logger:   logger,
	}, nil
}

func (e *EmailsPublisher) SetupExchangeAndQueue(exchange, queueName, bindingKey, consumerTag string) error {
	e.logger.Infof("Declaring exchange: %s", exchange)
	err := e.amqpChan.ExchangeDeclare(
		exchange,
		exchangeKind,
		exchangeDurable,
		exchangeAutoDelete,
		exchangeInternal,
		exchangeNoWait,
		nil,
	)

	if err != nil {
		return errors.Wrap(err, "e.amqpChan.ExchangeDeclare")
	}

	queue, err := e.amqpChan.QueueDeclare(
		queueName,
		queryDurable,
		queryAutoDelete,
		queryExclusive,
		queryNoWait,
		nil,
	)

	if err != nil {
		return errors.Wrap(err, "e.amqpChan.QueueDeclare")
	}

	e.logger.Infof("Declared queue, binding to exchange: Queue: %v, messageCount: %v, consumerCount: %v, exchange: %v, bindingKey: %v", queue.Name, queue.Messages, queue.Consumers, exchange, bindingKey)
	err = e.amqpChan.QueueBind(
		queue.Name,
		bindingKey,
		exchange,
		queryNoWait,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "e.amqpChan.QueueBind")
	}

	e.logger.Infof("Queue bound to exchange, starting to consumer from queue, consumerTag: %v", consumerTag)

	return nil
}

// Close messages channel
func (e *EmailsPublisher) CloseMessagesChannel() {
	if err := e.amqpChan.Close(); err != nil {
		e.logger.Errorf("e.amqpChan.Close: %v", err)
	}
}

// Publish message
func (e *EmailsPublisher) Publish(body []byte, contentType string) error {
	e.logger.Infof("Publishing message to exchange: %s, RoutingKey: %s", e.cfg.RabbitMQ.Exchange, e.cfg.RabbitMQ.RoutingKey)

	if err := e.amqpChan.Publish(
		e.cfg.RabbitMQ.Exchange,
		e.cfg.RabbitMQ.RoutingKey,
		publishNoWait,
		publishMandatory,
		amqp.Publishing{
			ContentType:  contentType,
			Body:         body,
			Timestamp:    time.Now(),
			MessageId:    uuid.New().String(),
			DeliveryMode: amqp.Persistent,
		},
	); err != nil {
		return errors.Wrap(err, "e.amqpChan.Publish")
	}

	e.logger.Infof("Message published to exchange: %s, RoutingKey: %s", e.cfg.RabbitMQ.Exchange, e.cfg.RabbitMQ.RoutingKey)

	publishedMessages.Inc()

	return nil
}
