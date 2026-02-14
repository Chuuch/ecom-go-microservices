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

	createProductTopic   = "create_product"
	createProductWorkers = 16
	updateProductTopic   = "update_product"
	updateProductWorkers = 16

	productsGropupID = "products_group"
)

// ProductsConsumerGroup
type ProductsConsumerGroup struct {
	Brokers    []string
	GroupID    string
	log        logger.Logger
	cfg        *config.Config
	productsUC product.UseCase
}

// NewProductsConsumerGroup constructor
func NewProductsConsumerGroup(brokers []string, groupID string, cfg *config.Config, productsUC product.UseCase, log logger.Logger) *ProductsConsumerGroup {
	return &ProductsConsumerGroup{
		Brokers:    brokers,
		GroupID:    groupID,
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
			Timeout: dialTimeout,
		},
	})
}

func (c *ProductsConsumerGroup) consumeCreateProduct(ctx context.Context, cancel context.CancelFunc, groupID string, topic string, workerNum int) {
	r := c.getNewReader(c.Brokers, topic, groupID)
	defer cancel()
	defer func() {
		if err := r.Close(); err != nil {
			c.log.Errorf("r.Close: %v", err)
			cancel()
		}
	}()

	c.log.Infof("Starting consume create product from topic: %s", topic)

	wg := &sync.WaitGroup{}
	for i := range workerNum {
		wg.Add(1)
		go c.createProductWorker(ctx, cancel, r, wg, i)
	}
	wg.Wait()
}

func (c *ProductsConsumerGroup) consumeUpdateProduct(ctx context.Context, cancel context.CancelFunc, groupID string, topic string, workerNum int) {
	r := c.getNewReader(c.Brokers, topic, groupID)
	defer cancel()
	defer func() {
		if err := r.Close(); err != nil {
			c.log.Errorf("r.Close: %v", err)
			cancel()
		}
	}()

	c.log.Infof("Starting consume update product from topic: %s", topic)

	wg := &sync.WaitGroup{}
	for i := range workerNum {
		wg.Add(1)
		go c.updateProductWorker(ctx, cancel, r, wg, i)
	}
	wg.Wait()
}

func (c *ProductsConsumerGroup) RunConsumers(ctx context.Context, cancel context.CancelFunc) {
	go c.consumeCreateProduct(ctx, cancel, productsGropupID, createProductTopic, createProductWorkers)
	go c.consumeUpdateProduct(ctx, cancel, productsGropupID, updateProductTopic, updateProductWorkers)
}
