package rabbitmq

import (
	"fmt"

	"github.com/Chuuch/ecom-microservices/config"
	"github.com/streadway/amqp"
)

// Initialize new RabbitMQ connection
func NewRabbitMQConnection(cfg *config.Config) (*amqp.Connection, error) {
	connAddr := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	)
	return amqp.Dial(connAddr)
}
