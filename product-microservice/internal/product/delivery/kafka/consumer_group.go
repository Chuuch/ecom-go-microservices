package kafka

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/chuuch/product-microservice/config"
	"github.com/chuuch/product-microservice/internal/models"
	"github.com/chuuch/product-microservice/internal/product"
	"github.com/chuuch/product-microservice/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

// ProductsConsumerGroup
type ProductsConsumerGroup struct {
	Brokers    []string
	GroupID    string
	log        logger.Logger
	cfg        *config.Config
	productsUC product.UseCase
	validate   *validator.Validate
}

// NewProductsConsumerGroup constructor
func NewProductsConsumerGroup(
	brokers []string,
	groupID string,
	cfg *config.Config,
	productsUC product.UseCase,
	log logger.Logger,
	validate *validator.Validate,
) *ProductsConsumerGroup {
	return &ProductsConsumerGroup{
		Brokers:    brokers,
		GroupID:    groupID,
		cfg:        cfg,
		productsUC: productsUC,
		log:        log,
		validate:   validate,
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

// Get new kafka.Writer
func (c *ProductsConsumerGroup) getNewWriter(topic string) *kafka.Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(c.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: -1,
		MaxAttempts:  maxAttempts,
		ReadTimeout:  writerReadTimeout,
		WriteTimeout: writerWriteTimeout,
		Logger:       kafka.LoggerFunc(c.log.Infof),
		ErrorLogger:  kafka.LoggerFunc(c.log.Errorf),
		Compression:  compress.Snappy,
	}

	return w
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

	w := c.getNewWriter(deadLetterQueueTopic)
	defer func() {
		if err := w.Close(); err != nil {
			c.log.Errorf("w.Close: %v", err)
			cancel()
		}
	}()

	c.log.Infof("Starting consume create product from topic: %s", topic)

	wg := &sync.WaitGroup{}
	for i := range workerNum {
		wg.Add(1)
		go c.createProductWorker(ctx, cancel, r, w, wg, i)
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

	w := c.getNewWriter(deadLetterQueueTopic)
	defer func() {
		if err := w.Close(); err != nil {
			c.log.Errorf("w.Close: %v", err)
			cancel()
		}
	}()

	c.log.Infof("Starting consume update product from topic: %s", topic)

	wg := &sync.WaitGroup{}
	for i := range workerNum {
		wg.Add(1)
		go c.updateProductWorker(ctx, cancel, r, w, wg, i)
	}
	wg.Wait()
}

func (c *ProductsConsumerGroup) publishErrorMessage(ctx context.Context, w *kafka.Writer, m kafka.Message, err error) error {
	errMsg := models.ErrorMessage{
		Offset:    m.Offset,
		Topic:     m.Topic,
		Partition: m.Partition,
		Error:     err.Error(),
		Time:      m.Time.UTC(),
	}

	errMsgBytes, err := json.Marshal(errMsg)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	return w.WriteMessages(ctx, kafka.Message{
		Value: errMsgBytes,
	})
}

func (c *ProductsConsumerGroup) RunConsumers(ctx context.Context, cancel context.CancelFunc) {
	go c.consumeCreateProduct(ctx, cancel, productsGropupID, createProductTopic, createProductWorkers)
	go c.consumeUpdateProduct(ctx, cancel, productsGropupID, updateProductTopic, updateProductWorkers)
}
