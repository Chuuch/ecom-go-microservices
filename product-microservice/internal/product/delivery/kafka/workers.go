package kafka

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/chuuch/product-microservice/internal/models"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/segmentio/kafka-go"
)

const (
	retryAttempts = 2
	retryDelay    = 1 * time.Second
)

func (c *ProductsConsumerGroup) createProductWorker(ctx context.Context, cancel context.CancelFunc, r *kafka.Reader, w *kafka.Writer, wg *sync.WaitGroup, workerID int) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ProductsConsumerGroup.createProductWorker")
	defer span.Finish()

	span.LogFields(log.String("ConsumerGroup", r.Config().GroupID))

	defer wg.Done()

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			c.log.Errorf("r.FetchMessage: %v", err)
			return
		}

		c.log.Infof(
			"WORKER: %v, message at topic/partition/offset: %v/%v/%v: %s = %s",
			workerID,
			m.Topic,
			m.Partition,
			m.Offset,
			string(m.Key),
			string(m.Value),
		)
		incomingMessages.Inc()

		var prod models.Product
		if err := json.Unmarshal(m.Value, &prod); err != nil {
			c.log.Errorf("json.Unmarshal: %v", err)
			errorMessages.Inc()
			continue
		}

		if err := c.validate.StructCtx(ctx, &prod); err != nil {
			errorMessages.Inc()
			c.log.Errorf("validate.StructCtx: %v", err)
			continue
		}

		if err := retry.Do(func() error {
			created, err := c.productsUC.CreateProduct(ctx, &prod)
			if err != nil {
				return err
			}
			c.log.Infof("Created product: %v", created.ProductID)
			return nil
		}, retry.Attempts(retryAttempts), retry.Delay(retryDelay), retry.Context(ctx)); err != nil {
			errorMessages.Inc()
			c.log.Errorf("retry.Do: %v", err)
			continue
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			errorMessages.Inc()
			c.log.Errorf("r.CommitMessages: %v", err)
			continue
		}

		successMessages.Inc()
	}
}

func (c *ProductsConsumerGroup) updateProductWorker(ctx context.Context, cancel context.CancelFunc, r *kafka.Reader, w *kafka.Writer, wg *sync.WaitGroup, workerID int) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ProductsConsumerGroup.updateProductWorker")
	defer span.Finish()

	span.LogFields(log.String("ConsumerGroup", r.Config().GroupID))

	defer wg.Done()
	defer cancel()

	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			c.log.Errorf("r.FetchMessage: %v", err)
			return
		}

		c.log.Infof(
			"WORKER: %v, message at topic/partition/offset: %v/%v/%v: %s = %s",
			workerID,
			m.Topic,
			m.Partition,
			m.Offset,
			string(m.Key),
			string(m.Value),
		)
		incomingMessages.Inc()

		var prod models.Product
		if err := json.Unmarshal(m.Value, &prod); err != nil {
			c.log.Errorf("json.Unmarshal: %v", err)
			errorMessages.Inc()
			continue
		}

		if err := retry.Do(func() error {
			updated, err := c.productsUC.UpdateProduct(ctx, &prod)
			if err != nil {
				return err
			}
			c.log.Infof("Updated product: %v", updated.ProductID)
			return nil
		}, retry.Attempts(retryAttempts), retry.Delay(retryDelay), retry.Context(ctx)); err != nil {
			errorMessages.Inc()
			c.log.Errorf("retry.Do: %v", err)
			if err := c.publishErrorMessage(ctx, w, m, err); err != nil {
				errorMessages.Inc()
				c.log.Errorf("productsConsumerGroup.updateProductWorker.publishErrorMessage: %v", err)
				continue
			}
			continue
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			errorMessages.Inc()
			c.log.Errorf("r.CommitMessages: %v", err)
			continue
		}

		successMessages.Inc()
	}
}
