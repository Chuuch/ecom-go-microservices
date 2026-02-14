package kafka

import (
	"context"

	"github.com/chuuch/product-microservice/config"
	"github.com/segmentio/kafka-go"
)

func NewKafkaConnection(cfg *config.Config) (*kafka.Conn, error) {
	return kafka.DialContext(context.Background(), "tcp", cfg.Kafka.Brokers[0])
}
