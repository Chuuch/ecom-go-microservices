package jaeger

import (
	"io"

	"github.com/chuuch/product-microservice/config"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"
	jaegerLog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

// Init Jaeger
func InitJaeger(c *config.Config) (opentracing.Tracer, io.Closer, error) {
	jaegerCfgInstance := jaegerConfig.Configuration{
		ServiceName: c.Jaeger.ServiceName,
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans:           c.Jaeger.LogSpans,
			LocalAgentHostPort: c.Jaeger.Host,
		},
	}

	return jaegerCfgInstance.NewTracer(
		jaegerConfig.Logger(jaegerLog.StdLogger),
		jaegerConfig.Metrics(metrics.NullFactory),
	)
}
