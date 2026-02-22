package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v5/middleware"
)

const (
	maxHeaderBytes = 1 << 20 // 1 MB
	stackSize      = 1 << 10 // 1 KB
	bodyLimit      = 2 << 20 // 2 MB
	readTimeout    = 15 * time.Second
	writeTimeout   = 15 * time.Second
	gzipLevel      = 5
)

func (a *App) runHttpServer() error {
	a.mapRoutes()
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", a.cfg.Http.Port),
		Handler:        a.echo,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
	return srv.ListenAndServe()
}

func (a *App) mapRoutes() {
	a.echo.Use(a.middlewareManager.RequestLoggerMiddleware)
	a.echo.Use(middleware.RecoverWithConfig(
		middleware.RecoverConfig{
			StackSize:         stackSize,
			DisableStackAll:   false,
			DisablePrintStack: false,
		},
	))
	a.echo.Use(middleware.RequestID())
	a.echo.Use(middleware.GzipWithConfig(
		middleware.GzipConfig{
			Level: gzipLevel,
		},
	))
	a.echo.Use(middleware.BodyLimit(bodyLimit))
}
