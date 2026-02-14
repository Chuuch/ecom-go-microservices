package kafka

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/chuuch/product-microservice/internal/models"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/segmentio/kafka-go"
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

		var prod models.Product
		if err := json.Unmarshal(m.Value, &prod); err != nil {
			c.log.Errorf("json.Unmarshal: %v", err)
			continue
		}

		created, err := c.productsUC.CreateProduct(ctx, &prod)
		if err != nil {
			errMsg := models.ErrorMessage{
				Offset:    m.Offset,
				Topic:     m.Topic,
				Partition: m.Partition,
				Error:     err.Error(),
				Time:      m.Time.UTC(),
			}

			errMsgBytes, err := json.Marshal(errMsg)
			if err != nil {
				c.log.Errorf("productsConsumerGroup.createProductWorker.json.Marshal: %v", err)
				continue
			}

			if err := w.WriteMessages(ctx, kafka.Message{
				Value: errMsgBytes,
			}); err != nil {
				c.log.Errorf("productsConsumerGroup.createProductWorker.w.WriteMessages: %v", err)
				continue
			}
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			c.log.Errorf("r.CommitMessages: %v", err)
			continue
		}

		c.log.Infof("WORKER: %v, created product: %v", workerID, created.ProductID)
	}
}

func (c *ProductsConsumerGroup) updateProductWorker(ctx context.Context, cancel context.CancelFunc, r *kafka.Reader, w *kafka.Writer, wg *sync.WaitGroup, workerID int) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ProductsConsumerGroup.updateProductWorker")
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

		var prod models.Product
		if err := json.Unmarshal(m.Value, &prod); err != nil {
			c.log.Errorf("json.Unmarshal: %v", err)
			continue
		}

		updated, err := c.productsUC.UpdateProduct(ctx, &prod)
		if err != nil {
			errMsg := models.ErrorMessage{
				Offset:    m.Offset,
				Topic:     m.Topic,
				Partition: m.Partition,
				Error:     err.Error(),
				Time:      m.Time.UTC(),
			}

			errMsgBytes, err := json.Marshal(errMsg)
			if err != nil {
				c.log.Errorf("productsConsumerGroup.updateProductWorker.json.Marshal: %v", err)
				continue
			}

			if err := w.WriteMessages(ctx, kafka.Message{
				Value: errMsgBytes,
			}); err != nil {
				continue
			}
		}

		if err := r.CommitMessages(ctx, m); err != nil {
			c.log.Errorf("r.CommitMessages: %v", err)
			continue
		}

		c.log.Infof("WORKER: %v, updated product: %v", workerID, updated.ProductID)
	}
}
