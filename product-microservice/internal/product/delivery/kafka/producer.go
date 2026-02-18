package kafka

import (
	"context"

	"github.com/chuuch/product-microservice/config"
	"github.com/chuuch/product-microservice/pkg/logger"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

// ProductsProducer interface
type ProductsProducer interface {
	PublishCreate(ctx context.Context, msgs ...kafka.Message) error
	PublishUpdate(ctx context.Context, msgs ...kafka.Message) error
	Close()
	Run()
	GetNewWriter(topic string) *kafka.Writer
}

type productsProducer struct {
	log          logger.Logger
	cfg          *config.Config
	createWriter *kafka.Writer
	updateWriter *kafka.Writer
}

func NewProductsProducer(log logger.Logger, cfg *config.Config) *productsProducer {
	return &productsProducer{
		log: log,
		cfg: cfg,
	}
}

// GetNewKafkaWriter Create new kafka writer
func (p *productsProducer) GetNewKafkaWriter(topic string) *kafka.Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(p.cfg.Kafka.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: -1,
		MaxAttempts:  maxAttempts,
		ReadTimeout:  writerReadTimeout,
		WriteTimeout: writerWriteTimeout,
		Logger:       kafka.LoggerFunc(p.log.Infof),
		ErrorLogger:  kafka.LoggerFunc(p.log.Errorf),
		Compression:  compress.Snappy,
	}
	return w
}
