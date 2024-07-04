package common

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"

	"disruptive/config"
)

// SetLogging initializeg the slog handler.
func SetLogging(level string) {
	var (
		handler  slog.Handler
		logLevel slog.Level
	)

	if level == "" {
		switch config.VARS.LoggingLevel {
		case "debug":
			logLevel = slog.LevelDebug
		case "warn":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		default:
			logLevel = slog.LevelInfo
		}
	} else {
		switch level {
		case "debug":
			logLevel = slog.LevelDebug
		case "warn":
			logLevel = slog.LevelWarn
		case "error":
			logLevel = slog.LevelError
		default:
			logLevel = slog.LevelInfo
		}
	}

	if isatty.IsTerminal(os.Stdout.Fd()) {
		sep := string(filepath.Separator)

		opts := &tint.Options{
			Level: logLevel,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey && len(groups) == 0 {
					return slog.Attr{}
				}

				if a.Key == slog.SourceKey {
					source := a.Value.Any().(*slog.Source)

					if strings.Contains(source.Function, "StartHTTPServer") ||
						strings.Contains(source.Function, "requestMiddleware") {
						return slog.Attr{}
					}

					source.Function = strings.Join(strings.Split(source.Function, sep)[1:], ".")
					source.File = filepath.Base(source.File)
				}
				return a
			},
			AddSource: true,
		}

		handler = tint.NewHandler(os.Stdout, opts)
	} else {
		opts := &slog.HandlerOptions{
			Level: logLevel,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey && len(groups) == 0 {
					return slog.Attr{}
				}
				return a
			},
		}

		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func requestMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)

		status := c.Response().Status

		httpErr := new(echo.HTTPError)
		if errors.As(err, &httpErr) {
			status = httpErr.Code
		}

		args := make([]any, 0, 8)
		args = append(args, slog.Int64("latency", int64(time.Since(start)/time.Millisecond)))
		args = append(args, slog.String("method", c.Request().Method))
		args = append(args, slog.String("path", c.Path()))
		args = append(args, slog.String("sid", c.Response().Header().Get(echo.HeaderXRequestID)))
		args = append(args, slog.Int("status", status))

		if uid, ok := c.Get("uid").(string); ok {
			args = append(args, slog.String("uid", uid))
		}

		slog.Info("request", args...)
		return err
	}
}
