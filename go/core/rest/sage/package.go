// Package sage manages the sage project
package sage

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	e "disruptive/rest/errors"
)

var (
	sageUser = []string{"sage.user"}
)

func getUUIDParam(c echo.Context, logCtx *slog.Logger, param string) (*slog.Logger, string, error) {
	fid := slog.String("fid", "rest.sage.getUUIDParam")

	if len(param) < 2 || param[0] != '_' {
		return nil, "", e.ErrBad(logCtx, fid, fmt.Sprintf("invalid param: %s", param))
	}

	k := param[1:]
	v := c.Param(param)
	if v == "" {
		return nil, "", e.ErrBad(logCtx, fid, fmt.Sprintf("invalid param: %s", param))
	}

	if _, err := uuid.Parse(v); err != nil {
		return nil, "", e.ErrBad(logCtx, fid, fmt.Sprintf("invalid param: %s uuid: %s", param, v))
	}

	logCtx = logCtx.With(k, v)
	return logCtx, v, nil
}
