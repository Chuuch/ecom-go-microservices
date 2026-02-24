package rabbitmq

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/chuuch/search-microservice/config"
	"github.com/chuuch/search-microservice/internal/product/domain"
	"github.com/chuuch/search-microservice/pkg/logger"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
)

type productConsumer struct {
	log            logger.Logger
	cfg            *config.Config
	amqpConn       *amqp091.Connection
	amqpChan       *amqp091.Channel
	productUsecase domain.ProductUsecase
	bulkIndexer    esutil.BulkIndexer
	esClient       *elasticsearch.Client
}

func NewProductConsumer(
	log logger.Logger,
	cfg *config.Config,
	amqpConn *amqp091.Connection,
	amqpChan *amqp091.Channel,
	productUsecase domain.ProductUsecase,
	esClient *elasticsearch.Client,
) *productConsumer {
	return &productConsumer{
		log:            log,
		cfg:            cfg,
		amqpConn:       amqpConn,
		amqpChan:       amqpChan,
		productUsecase: productUsecase,
		esClient:       esClient,
	}
}

func (c *productConsumer) ConsumeIndexDeliveries(
	ctx context.Context,
	deliveries <-chan amqp091.Delivery,
	workerID int,
) func() error {
	return func() error {
		c.log.Infof("starting to consume index deliveries for worker: %d", workerID)

		for {
			select {
			case <-ctx.Done():
				c.log.Infof("context done, stopping to consume index deliveries")
				return ctx.Err()
			case delivery, ok := <-deliveries:
				if !ok {
					c.log.Infof("deliveries channel closed, stopping to consume index deliveries")
					return nil
				}
				if err := c.bulkIndexProduct(ctx, delivery); err != nil {
					c.log.Errorf("failed to index product: %v", err)
					return err
				}
			}
		}
	}
}

func (c *productConsumer) indexProduct(ctx context.Context, msg amqp091.Delivery) error {
	var product domain.Product

	if err := json.Unmarshal(msg.Body, &product); err != nil {
		c.log.Errorf("failed to unmarshal product: %v", err)
		return msg.Reject(true)
	}
	if err := c.productUsecase.Index(ctx, product); err != nil {
		c.log.Errorf("failed to index product: %v", err)
		return msg.Reject(true)
	}
	c.log.Infof("product indexed successfully: %s", product.ID)
	return msg.Ack(true)
}

func (c *productConsumer) bulkIndexProduct(ctx context.Context, msg amqp091.Delivery) error {
	var product domain.Product

	if err := json.Unmarshal(msg.Body, &product); err != nil {
		c.log.Errorf("failed to unmarshal produt: %v", err)
		return msg.Reject(true)
	}

	if err := c.bulkIndexer.Add(
		ctx,
		esutil.BulkIndexerItem{
			Index:      c.cfg.ElasticMapping.ProductsIndex.Name,
			Action:     "create",
			DocumentID: product.ID,
			Body:       bytes.NewReader(msg.Body),
			OnSuccess: func(
				ctx context.Context,
				item esutil.BulkIndexerItem,
				response esutil.BulkIndexerResponseItem,
			) {
				c.log.Infof("failed to index product: %v", response.Error)
			},
		},
	); err != nil {
		c.log.Errorf("failed to add product to bulk indexer: %v", err)
		return msg.Reject(true)
	}

	c.log.Infof("product added to bulk indexer: %s", product.ID)
	return msg.Ack(true)
}

func (c *productConsumer) InitBulkIndexer() error {
	bulkIndexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		NumWorkers:    c.cfg.RabbitMQ.BulkIndexer.NumWorkers,
		FlushBytes:    c.cfg.RabbitMQ.BulkIndexer.FlushBytes,
		FlushInterval: time.Duration(c.cfg.RabbitMQ.BulkIndexer.FlushIntervalSeconds) * time.Second,
		Client:        c.esClient,
		OnError: func(ctx context.Context, err error) {
			c.log.Errorf("failed to initialize bulk indexer: %v", err)
		},
		OnFlushStart: func(ctx context.Context) context.Context {
			c.log.Infof("flush started")
			return ctx
		},
		OnFlushEnd: func(ctx context.Context) {
			c.log.Infof("flush completed")
		},
		Index:   c.cfg.ElasticMapping.ProductsIndex.Name,
		Human:   true,
		Pretty:  true,
		Timeout: time.Duration(c.cfg.RabbitMQ.BulkIndexer.TimeoutMilliseconds) * time.Second,
	})
	if err != nil {
		return errors.Wrap(err, "failed to initialize bulk indexer")
	}
	c.bulkIndexer = bulkIndexer
	c.log.Infof("bulk indexer initialized")
	return nil
}

func (c *productConsumer) Close(ctx context.Context) {
	if err := c.bulkIndexer.Close(ctx); err != nil {
		c.log.Errorf("failed to close bulk indexer: %v", err)
	}
	c.log.Infof("bulk indexer closed")
}
