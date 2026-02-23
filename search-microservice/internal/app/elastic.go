package app

import (
	"context"
	"time"

	"github.com/avast/retry-go"
	"github.com/chuuch/search-microservice/pkg/elastic"
	"github.com/chuuch/search-microservice/pkg/esclient"
)

func (a *App) initElasticSearchClient(ctx context.Context) error {
	retryOptions := []retry.Option{
		retry.Attempts(5),
		retry.Delay(time.Duration(1500) * time.Microsecond),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, err error) {
			a.log.Errorf("ElasticSearch client initialization failed: %v", err)
		}),
	}

	return retry.Do(func() error {
		esClient, err := elastic.NewElasticSearch(a.cfg)
		if err != nil {
			return err
		}
		a.elasticClient = esClient

		elasticResponse, err := esclient.Info(ctx, a.elasticClient)
		if err != nil {
			return err
		}

		defer elasticResponse.Body.Close()

		a.log.Info("ElasticSearch is reachable %s", elasticResponse.String())
		return nil
	}, retryOptions...)
}
