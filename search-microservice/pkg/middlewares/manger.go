package middlewares

import (
	"strings"
	"time"

	"github.com/chuuch/search-microservice/config"
	"github.com/chuuch/search-microservice/pkg/logger"
	"github.com/labstack/echo/v5"
)

type MiddlewareMetricsCb func(err error)

type MiddlewareManager interface {
	RequestLoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc
}

type middlewareManager struct {
	log       logger.Logger
	cfg       *config.Config
	metricsCb MiddlewareMetricsCb
}

func NewMiddlewareManager(
	log logger.Logger,
	cfg *config.Config,
	metricsCb MiddlewareMetricsCb,
) *middlewareManager {
	return &middlewareManager{
		log:       log,
		cfg:       cfg,
		metricsCb: metricsCb,
	}
}

func (mw *middlewareManager) RequestLoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx *echo.Context) error {
		start := time.Now()
		err := next(ctx)

		req := ctx.Request()
		res := ctx.Response()
		echoRes, status := echo.ResolveResponseStatus(res, err)
		size := int64(-1)
		if echoRes != nil {
			size = echoRes.Size
		}
		s := time.Since(start)

		if !mw.checkIgnoredURI(req.RequestURI, mw.cfg.Http.IgnoredURIs) {
			mw.log.Info("request", "method", req.Method, "uri", req.RequestURI, "status", status, "size", size, "duration", s)
		}

		if mw.metricsCb != nil {
			mw.metricsCb(err)
		}

		return err
	}
}

func (mw *middlewareManager) checkIgnoredURI(requestURI string, uriList []string) bool {
	for _, uri := range uriList {
		if strings.Contains(requestURI, uri) {
			return true
		}
	}
	return false
}
