package grpc

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	incommingMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_incoming_grpc_messages_total",
		Help: "Total number of incoming gRPC messages",
	})

	successMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_success_incoming_grpc_messages_total",
		Help: "Total number of successful incoming gRPC messages",
	})

	errorMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_error_incoming_grpc_messages_total",
		Help: "Total number of failed incoming gRPC messages",
	})
)
