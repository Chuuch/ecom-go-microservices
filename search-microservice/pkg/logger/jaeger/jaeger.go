package jaeger

import (
	"io"

	"github.com/chuuch/search-microservice/config"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	jaegerLog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

// Init Jaeger
func InitJaeger(cfg *config.Config) (opentracing.Tracer, io.Closer, error) {
	jaegerCfg := jaegerConfig.Configuration{
		ServiceName: cfg.Jaeger.ServiceName,
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans:          cfg.Jaeger.LogSpans,
			CollectorEndpoint: cfg.Jaeger.Host,
		},
	}

	return jaegerCfg.NewTracer(
		jaegerConfig.Logger(jaegerLog.StdLogger),
		jaegerConfig.Metrics(metrics.NullFactory),
	)
}
