package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) StartHTTP() {
	s.echo.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"message": "OK"})
	})
	s.echo.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	//middlewares
	s.mapRoutes()

	go func() {
		s.logger.Infof("HTTP server is running on port %s", s.cfg.Http.Port)
		s.echo.Server.ReadTimeout = time.Second * s.cfg.Http.ReadTimeout
		s.echo.Server.WriteTimeout = time.Second * s.cfg.Http.WriteTimeout
		if err := s.echo.Start(s.cfg.Http.Port); err != nil {
			s.logger.Fatal(err)
		}
	}()
}

func (s *Server) mapRoutes() {
	s.echo.Use(middleware.RequestLogger())
	s.echo.Use(middleware.CORSWithConfig(
		middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
			AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		},
	))
	s.echo.Use(middleware.RecoverWithConfig(
		middleware.RecoverConfig{
			DisablePrintStack: true,
			DisableStackAll:   true,
		},
	))
	s.echo.Use(middleware.RequestID())
	s.echo.Use(middleware.GzipWithConfig(
		middleware.GzipConfig{
			Level: 5,
			Skipper: func(c echo.Context) bool {
				return strings.Contains(c.Request().URL.Path, "swagger")
			},
		},
	))
}
