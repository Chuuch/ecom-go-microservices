package kafka

import (
	"context"
	"sync"
	"time"

	"github.com/chuuch/product-microservice/config"
	"github.com/chuuch/product-microservice/internal/product"
	"github.com/chuuch/product-microservice/pkg/logger"
	"github.com/segmentio/kafka-go"
)

const (
	minBytes               = 10e3            // fetch at least 10KB of messages
	maxBytes               = 10e6            // fetch at most 10MB of messages
	queueCapacity          = 100             // internal message buffer size in kafka.Reader
	heartbeatInterval      = 3 * time.Second // hearbeat interval to the group coordinator
	commitInterval         = 0               // commit offsets synchronously on every message
	partitionWatchInterval = 5 * time.Second // how often to watch fo partition changes
	maxAttempts            = 5               // maximum number of attempts for transient errors
	dialTimeout            = 3 * time.Minute // timeout for connecting to the broker
)

// ProductsConsumerGroup
type ProductsConsumerGroup struct {
	Brokers    []string
	GroupID    string
	Topic      string
	log        logger.Logger
	cfg        config.Config
	productsUC product.UseCase
}

// NewProductsConsumerGroup constructor
func NewProductsConsumerGroup(brokers []string, groupID string, topic string, cfg config.Config, productsUC product.UseCase, log logger.Logger) *ProductsConsumerGroup {
	return &ProductsConsumerGroup{
		Brokers:    brokers,
		GroupID:    groupID,
		Topic:      topic,
		cfg:        cfg,
		productsUC: productsUC,
		log:        log,
	}
}

// Get new kafka.Reader
func (c *ProductsConsumerGroup) getNewReader(kafkaURL []string, topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:                kafkaURL,
		GroupID:                groupID,
		Topic:                  topic,
		MinBytes:               minBytes,
		MaxBytes:               maxBytes,
		QueueCapacity:          queueCapacity,
		HeartbeatInterval:      heartbeatInterval,
		CommitInterval:         commitInterval,
		PartitionWatchInterval: partitionWatchInterval,
		MaxAttempts:            maxAttempts,
		Logger:                 kafka.LoggerFunc(c.log.Infof),
		ErrorLogger:            kafka.LoggerFunc(c.log.Errorf),
		Dialer: &kafka.Dialer{
			Timeout:         dialTimeout,
			LocalAddr:       nil,
			DualStack:       false,
			FallbackDelay:   0,
			KeepAlive:       0,
			Resolver:        nil,
			TLS:             nil,
			SASLMechanism:   nil,
			TransactionalID: "",
		},
	})
}

func (c *ProductsConsumerGroup) consumeCreateProduct(ctx context.Context, cancel context.CancelFunc, topic string, workerNum int) {
	r := c.getNewReader(c.Brokers, topic, c.GroupID)
	defer cancel()
	defer func() {
		if err := r.Close(); err != nil {
			c.log.Errorf("r.Close: %v", err)
			cancel()
		}
	}()

	c.log.Infof("Starting consume create product from topic: %s", topic)

	wg := &sync.WaitGroup{}
	for range workerNum {
		wg.Add(1)
		// TODO: Start a new worker
	}
	wg.Wait()
}

func (c *ProductsConsumerGroup) consumeUpdateProduct(ctx context.Context, cancel context.CancelFunc, topic string, workerNum int) {
	r := c.getNewReader(c.Brokers, topic, c.GroupID)
	defer cancel()
	defer func() {
		if err := r.Close(); err != nil {
			c.log.Errorf("r.Close: %v", err)
			cancel()
		}
	}()

	c.log.Infof("Starting consume update product from topic: %s", topic)

	wg := &sync.WaitGroup{}
	for range workerNum {
		wg.Add(1)
		// TODO: Start a new worker
	}
	wg.Wait()
}

func (c *ProductsConsumerGroup) RunConsumers(ctx context.Context, cancel context.CancelFunc) {
	go c.consumeCreateProduct(ctx, cancel, c.Topic, 16)
	go c.consumeUpdateProduct(ctx, cancel, c.Topic, 16)
}
