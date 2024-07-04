package common

import (
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"disruptive/config"
)

// StartHTTPServer starts an HTTP server and starts listening.
func StartHTTPServer(listener net.Listener, routes ...func(e *echo.Echo)) {
	e := echo.New()
	e.Logger.SetOutput(io.Discard)
	e.HideBanner = true
	e.Use(requestMiddleware, middleware.CORS(), middleware.Gzip(), middleware.RequestID())

	for _, r := range routes {
		r(e)
	}

	slog.Info("HTTP server started", "port", config.VARS.Port, "loglevel", config.VARS.LoggingLevel)

	if err := e.Server.Serve(listener); err != nil {
		slog.Error("failed to serve HTTP", "error", err.Error())
		os.Exit(1)
	}
}
