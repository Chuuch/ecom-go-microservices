package kafka

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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

	createProductTopic   = "create_product" // topic to consume create product messages
	createProductWorkers = 16               // number of workers to consume create product messages
	updateProductTopic   = "update_product" // topic to consume update product messages
	updateProductWorkers = 16               // number of workers to consume update product messages

	productsGropupID     = "products_group"    // group id for products
	deadLetterQueueTopic = "dead_letter_queue" // topic to consume dead letter queue messages

	writerReadTimeout  = 10 * time.Second // timeout for reading from kafka writer
	writerWriteTimeout = 10 * time.Second // timeout for writing to kafka writer
)

var (
	incomingMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_incoming_kafka_messages_total",
		Help: "Total number of incoming kafka messages",
	})

	successMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_success_incoming_kafka_messages_total",
		Help: "Total number of successful incoming kafka messages",
	})

	errorMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_error_incoming_kafka_messages_total",
		Help: "Total number of failed incoming kafka messages",
	})
)
