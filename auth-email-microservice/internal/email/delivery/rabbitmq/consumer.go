package rabbitmq

import (
	"context"

	"github.com/Chuuch/ecom-microservices/internal/email"
	"github.com/Chuuch/ecom-microservices/pkg/logger"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/streadway/amqp"
)

const (
	exchangeKind       = "direct"
	exchangeDurable    = true  // survives broker restart
	exchangeAutoDelete = false // don't delete when unused
	exchangeInternal   = false // exchange is usable by publishers
	exchangeNoWait     = false // wait for declare confirmation

	queryDurable    = true  // queue persists across restarts
	queryAutoDelete = false // don't auto delete
	queryExclusive  = false // not tied to a single connection
	queryNoWait     = false

	prefetchCount  = 1     // one unacked at a time per consumer
	prefetchSize   = 0     // size-based limit disabled
	prefetchGlobal = false // apply per-consumer not globally

	consumerAutoAck   = false // we manually ack/reject
	consumerExclusive = false // allow multiple consumers
	consumeNoLocal    = false
	consumeNoWait     = false //

	publishNoWait    = false // fail if not routed
	publishMandatory = false // fail if no consumers
)

var (
	IncomingMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "email_incoming_rabbitmq_messages_total",
		Help: "The total number of incoming RabbitMQ messages",
	})
	successMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "email_success_rabbitmq_messages_total",
		Help: "The total number of successful RabbitMQ messages",
	})
	failureMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "email_failure_rabbitmq_messages_total",
		Help: "The total number of failed RabbitMQ messages",
	})
)

// Email consumer
type EmailConsumer struct {
	amqpConn *amqp.Connection
	logger   logger.Logger
	emailUC  email.EmailUseCase
}

// New Email consumer
func NewEmailConsumer(amqpConn *amqp.Connection, logger logger.Logger, emailUC email.EmailUseCase) *EmailConsumer {
	return &EmailConsumer{
		amqpConn: amqpConn,
		logger:   logger,
		emailUC:  emailUC,
	}
}

// Consumer messages
func (c *EmailConsumer) CreateChannel(exchangeName, queueName, bindingKey, consumerTag string) (*amqp.Channel, error) {
	ch, err := c.amqpConn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "amqpConn.Channel")
	}

	c.logger.Infof("Declaring exchange: %s", exchangeName)
	err = ch.ExchangeDeclare(
		exchangeName,
		exchangeKind,
		exchangeDurable,
		exchangeAutoDelete,
		exchangeInternal,
		exchangeNoWait,
		nil,
	)

	if err != nil {
		return nil, errors.Wrap(err, "ch.ExchangeDeclare")
	}

	queue, err := ch.QueueDeclare(
		queueName,
		queryDurable,
		queryAutoDelete,
		queryExclusive,
		queryNoWait,
		nil,
	)

	if err != nil {
		return nil, errors.Wrap(err, "ch.QueueDeclare")
	}

	c.logger.Infof(
		"Declared queue, binding to exchange: Queue: %v, messageCount: %v, "+
			"consumerCount: %v, exchange: %v, bindingKey: %v",
		queue.Name,
		queue.Messages,
		queue.Consumers,
		exchangeName,
		bindingKey,
	)

	err = ch.QueueBind(
		queue.Name,
		bindingKey,
		exchangeName,
		queryNoWait,
		nil,
	)

	if err != nil {
		return nil, errors.Wrap(err, "ch.QueueBind")
	}

	c.logger.Infof("Queue bound to exchange, starting to consume from queue, consumerTag: %v", consumerTag)

	err = ch.Qos(
		prefetchCount,
		prefetchSize,
		prefetchGlobal,
	)
	if err != nil {
		return nil, errors.Wrap(err, "ch.Qos")
	}

	return ch, nil
}

func (c *EmailConsumer) worker(ctx context.Context, messages <-chan amqp.Delivery) {

	for delivery := range messages {
		span, ctx := opentracing.StartSpanFromContext(ctx, "EmailConsumer.worker")

		c.logger.Infof("processDeliveries deliveryTag: %v", delivery.DeliveryTag)

		IncomingMessages.Inc()

		err := c.emailUC.SendEmail(ctx, delivery)
		if err != nil {
			if err := delivery.Reject(false); err != nil {
				c.logger.Errorf("delivery.Reject: %v", err)
			}
			c.logger.Errorf("emailUC.SendEmail, Failed to process delivery: %v", err)
			failureMessages.Inc()
			span.Finish()
		} else {
			err = delivery.Ack(false)
			if err != nil {
				c.logger.Errorf("delivery.Ack: %v", err)
			}
			successMessages.Inc()
			span.Finish()
		}
	}
	c.logger.Infof("Deliveries channel closed")
}

// Start new rabbitmq consumer
func (c *EmailConsumer) StartConsumer(workerPoolSize int, exchangeName, queueName, bindingKey, consumerTag string) error {
	ch, err := c.CreateChannel(exchangeName, queueName, bindingKey, consumerTag)
	if err != nil {
		return err
	}
	defer ch.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deliveries, err := ch.Consume(
		queueName,
		consumerTag,
		consumerAutoAck,
		consumerExclusive,
		consumeNoLocal,
		consumeNoWait,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "ch.Consume")
	}

	for range workerPoolSize {
		go c.worker(ctx, deliveries)
	}

	// Shutdown gracefully
	chanErr := <-ch.NotifyClose(make(chan *amqp.Error))
	c.logger.Errorf("ch.NotifyClose: %v", chanErr)

	return chanErr
}
