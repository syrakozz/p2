package api

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"

	"disruptive/config"
)

func getInfo(c echo.Context) error {
	slog.Info("info", "fid", "rest.getInfo", "sid", c.Response().Header().Get(echo.HeaderXRequestID), "env", config.VARS.Env)

	return c.JSON(http.StatusOK, map[string]any{
		"port":           config.VARS.Port,
		"env":            config.VARS.Env,
		"service":        filepath.Base(os.Args[0]),
		"user_agent":     config.VARS.UserAgent,
		"version":        strings.Split(config.VARS.BuildTag, "/")[1], // ok to panic
		"build_image":    config.VARS.BuildImage,
		"build_commit":   config.VARS.BuildCommit,
		"build_datetime": config.VARS.BuildDateTime,
		"build_tag":      config.VARS.BuildTag,
		"logging_level":  config.VARS.LoggingLevel,
	})
}
