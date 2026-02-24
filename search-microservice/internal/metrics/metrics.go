package metrics

import (
	"github.com/chuuch/search-microservice/config"
	"github.com/prometheus/client_golang/prometheus"
)

type SearchMicroserviceMetrics struct {
	SuccessHttpRequestsTotal prometheus.Counter
	ErrorHttpRequestsTotal   prometheus.Counter

	HttpSuccessIndexAsyncRequests prometheus.Counter
	HttpSuccessIndexRequests      prometheus.Counter
	HttpSuccessSearchRequests     prometheus.Counter

	RabbitMQSuccessBatchInsertMessages prometheus.Counter
}

func NewSearchMicroserviceMetrics(cfg *config.Config) *SearchMicroserviceMetrics {

	successCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "search_microservice_success_http_requests_total",
		Help: "The total number of successful HTTP requests",
	})
	errorCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "search_microservice_error_http_requests_total",
		Help: "The total number of error HTTP requests",
	})
	HttpSuccessIndexAsyncRequestsCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "search_microservice_http_success_index_async_requests_total",
		Help: "The total number of successful index async HTTP requests",
	})
	HttpSuccessIndexRequestsCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "search_microservice_http_success_index_requests_total",
		Help: "The total number of successful index HTTP requests",
	})
	HttpSuccessSearchRequestsCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "search_microservice_http_success_search_requests_total",
		Help: "The total number of successful batch insert messages",
	})
	RabbitMQSuccessBatchInsertMessagesCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "search_microservice_rabbitmq_success_batch_insert_messages_total",
		Help: "The total number of successful batch insert messages",
	})

	prometheus.MustRegister(successCounter)
	prometheus.MustRegister(errorCounter)
	prometheus.MustRegister(HttpSuccessIndexAsyncRequestsCounter)
	prometheus.MustRegister(HttpSuccessIndexRequestsCounter)
	prometheus.MustRegister(HttpSuccessSearchRequestsCounter)
	prometheus.MustRegister(RabbitMQSuccessBatchInsertMessagesCounter)

	return &SearchMicroserviceMetrics{
		SuccessHttpRequestsTotal:           successCounter,
		ErrorHttpRequestsTotal:             errorCounter,
		HttpSuccessIndexAsyncRequests:      HttpSuccessIndexAsyncRequestsCounter,
		HttpSuccessIndexRequests:           HttpSuccessIndexRequestsCounter,
		HttpSuccessSearchRequests:          HttpSuccessSearchRequestsCounter,
		RabbitMQSuccessBatchInsertMessages: RabbitMQSuccessBatchInsertMessagesCounter,
	}
}
