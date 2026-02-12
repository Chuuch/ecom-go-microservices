package mongodb

import (
	"context"
	"time"

	"github.com/chuuch/product-microservice/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	connectTimeout  = 30 * time.Second
	maxConnIdleTime = 3 * time.Minute
	minPoolSize     = 20
	maxPoolSize     = 300
)

// NewMongoDBConn Create new MongoDB client
func NewMongoDBConn(ctx context.Context, cfg *config.Config) (*mongo.Client, error) {
	opts := options.Client().
		ApplyURI(cfg.MongoDB.URI).
		SetAuth(options.Credential{
			Username: cfg.MongoDB.User,
			Password: cfg.MongoDB.Password,
		}).
		SetConnectTimeout(connectTimeout).
		SetMaxConnIdleTime(maxConnIdleTime).
		SetMinPoolSize(minPoolSize).
		SetMaxPoolSize(maxPoolSize)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}
