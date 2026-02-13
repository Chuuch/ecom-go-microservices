package interceptors

import (
	"context"
	"time"

	"github.com/chuuch/product-microservice/config"
	"github.com/chuuch/product-microservice/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// InterceptorManager
type InterceptorManager struct {
	logger logger.Logger
	cfg    *config.Config
}

// InterceptorManager constructor
func NewInterceptorManager(logger logger.Logger, cfg *config.Config) *InterceptorManager {
	return &InterceptorManager{
		logger: logger,
		cfg:    cfg,
	}
}

// Logger Interceptor
func (im *InterceptorManager) Logger(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ctx)
	reply, err := handler(ctx, req)
	im.logger.Infof("METHOD: %s, REQUEST: %v, RESPONSE: %v, ERROR: %v, TIME: %s, METADATA: %v", info.FullMethod, req, reply, err, time.Since(start), md)
	return reply, err
}
