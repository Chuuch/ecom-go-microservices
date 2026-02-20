package interceptors

import (
	"context"
	"net/http"
	"time"

	"github.com/Chuuch/ecom-microservices/config"
	grpcerrors "github.com/Chuuch/ecom-microservices/pkg/grpc_errors"
	"github.com/Chuuch/ecom-microservices/pkg/logger"
	"github.com/Chuuch/ecom-microservices/pkg/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// InterceptorManager

type InterceptorManager struct {
	logger logger.Logger
	cfg    *config.Config
	mtr    metric.Metrics
}

// InterceptorManager constructor
func NewInterceptorManager(logger logger.Logger, cfg *config.Config, mtr metric.Metrics) *InterceptorManager {
	return &InterceptorManager{
		logger: logger,
		cfg:    cfg,
		mtr:    mtr,
	}
}

// Logger interceptors
func (im *InterceptorManager) Logger(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ctx)
	reply, err := handler(ctx, req)

	im.logger.Infof("METHOD: %s, REQUEST: %v, RESPONSE: %v ERROR: %v, TIME: %s, METADATA: %v", info.FullMethod, req, reply, err, time.Since(start), md)

	return reply, err
}

func (im *InterceptorManager) Metrics(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	var status int = http.StatusOK
	if err != nil {
		status = grpcerrors.MapGRPCErrStatusCodeToHttpStatus(grpcerrors.ParseGRPCError(err))
	}
	im.mtr.ObserveResponseTime(status, info.FullMethod, info.FullMethod, float64(time.Since(start).Seconds()))
	im.mtr.IncHits(status, info.FullMethod, info.FullMethod)
	return resp, err

}
