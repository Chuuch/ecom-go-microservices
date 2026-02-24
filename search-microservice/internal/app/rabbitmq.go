package app

import (
	"context"
	"time"

	"github.com/avast/retry-go"
	"github.com/chuuch/search-microservice/pkg/rabbitmq"
)

func (a *App) initRabbitMQ(ctx context.Context) error {
	retryOptions := []retry.Option{
		retry.Attempts(5),
		retry.Delay(time.Duration(1500) * time.Microsecond),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, err error) {
			a.log.Errorf("RabbitMQ client initialization failed: %v", err)
		}),
	}

	return retry.Do(func() error {
		amqpConn, err := rabbitmq.NewRabbitMQConnection(a.cfg)
		if err != nil {
			return err
		}
		a.amqpConn = amqpConn

		amqpChan, err := amqpConn.Channel()
		if err != nil {
			return err
		}
		a.amqpChan = amqpChan

		if err := a.amqpChan.Qos(1, 0, true); err != nil {
			a.log.Error("failed to set QoS: %v", err)
			return err
		}

		a.log.Info("RabbigMQ client initalized")
		return nil
	}, retryOptions...)
}

func (a *App) initRabbitMQPublisher(ctx context.Context) error {
	retryOptions := []retry.Option{
		retry.Attempts(5),
		retry.Delay(time.Duration(1500) * time.Millisecond),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, err error) {
			a.log.Errorf("RabbitMQ publisher initialization failed: %v", err)
		}),
	}

	return retry.Do(func() error {
		amqpPublisher, err := rabbitmq.NewPublisher(a.cfg, a.log)
		if err != nil {
			return err
		}
		a.amqpPublisher = amqpPublisher

		a.log.Info("RabbitMQ publisher initialized")
		return nil
	}, retryOptions...)
}

func (a *App) closeRabbitMQ() error {
	if a.amqpConsumeChan != nil {
		_ = a.amqpConsumeChan.Close()
	}
	if err := a.amqpChan.Close(); err != nil {
		return err
	}

	if err := a.amqpConn.Close(); err != nil {
		return err
	}
	return nil
}
