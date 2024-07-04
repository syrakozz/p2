// Package stabilityai interfaces directly with the Stability AI APIs.
package stabilityai

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/lib/stabilityai"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// PostTTI starts a StableDiffision text-to-image process.
func PostTTI(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.lib.stabilityai.PostTTI")

	engine := c.QueryParam("engine")
	if engine != "" {
		if _, ok := stabilityai.Engines[engine]; !ok {
			return e.ErrBad(logCtx, fid, "invalid engine")
		}
	}

	req := stabilityai.Request{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if len(req.TextPrompts) < 1 {
		return e.ErrBad(logCtx, fid, "invalid prompt")
	}

	if req.TextPrompts[0].Text == "" {
		return e.ErrBad(logCtx, fid, "invalid prompt")
	}

	res, err := stabilityai.PostTTI(ctx, logCtx, engine, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process text-to-image")
	}

	return c.Blob(http.StatusOK, "image/png", res)
}
