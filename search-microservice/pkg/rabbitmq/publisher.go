package rabbitmq

import (
	"context"

	"github.com/chuuch/search-microservice/config"
	"github.com/chuuch/search-microservice/pkg/logger"
	"github.com/rabbitmq/amqp091-go"
)

type AmqpPublisher interface {
	PublishWithContext(ctx context.Context, exhange, key string, mandatoy, immediate bool, msg amqp091.Publishing) error
	Publish(ctx context.Context, exchange, key string, msg amqp091.Publishing) error
	Close() error
}

type publisher struct {
	amqpConn *amqp091.Connection
	amqpChan *amqp091.Channel
	log      logger.Logger
}


func NewPublisher(cfg *config.Config, log logger.Logger) (*publisher, error) {
	conn, err := NewRabbitMQConnection(cfg)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &publisher{
		amqpConn: conn,
		amqpChan: channel,
		log: log,
	}, nil
}

func (p *publisher) Close() {
	if err := p.amqpChan.Close(); err != nil {
		p.log.Errorf("failed to close amqp channel: %v", err)
	}
	if err := p.amqpConn.Close(); err != nil {
		p.log.Errorf("failed to close amqp connection: %v", err)
	}
}

func (p *publisher) PublishWithContext(
	ctx context.Context,
	exchange string,
	key string,
	mandatory bool,
	immediate bool,
	msg amqp091.Publishing) error {
	if err := p.amqpChan.PublishWithContext(
		ctx,
		exchange,
		key,
		mandatory,
		immediate,
		msg); err != nil {
			p.log.Error("failed to publish message: %v", err)
			return err
		}
		return nil
}

func (p *publisher) Publish(ctx context.Context, exchange, key string, msg amqp091.Publishing) error {
	if err := p.amqpChan.PublishWithContext(ctx, exchange, key, false, false, msg); err != nil {
		p.log.Error("failed to pubilsh message: %v", err)
		return err
	}
	return nil
}