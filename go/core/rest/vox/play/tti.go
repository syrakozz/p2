package play

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"disruptive/lib/configs"
	"disruptive/lib/stabilityai"
	"disruptive/pkg/vox/characters/play"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetTTI is the REST API to get an image based on last memory.
func GetTTI(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.GetTTI")

	profileID := c.Param("profile_id")
	characterVersion := c.Param("character_version")
	engine := c.QueryParam("engine")
	width := c.QueryParam("width")
	height := c.QueryParam("height")
	entries := c.QueryParam("entries")

	if engine != "" {
		if _, ok := stabilityai.Engines[engine]; !ok {
			return e.ErrBad(logCtx, fid, "invalid engine")
		}
	}

	var (
		h, w int
		err  error
	)

	w, err = strconv.Atoi(width)
	if err != nil {
		w = 0
	}

	h, err = strconv.Atoi(height)
	if err != nil {
		h = 0
	}

	n, err := strconv.Atoi(entries)
	if err != nil || n == 0 {
		n = 3
	}

	image, err := play.TTI(ctx, logCtx, profileID, characterVersion, engine, w, h, n)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process text-to-image")
	}

	return c.Blob(http.StatusOK, "image/png", image)
}

// GetTTIStyles is the REST API to get the available image styles.
func GetTTIStyles(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.GetTTIStyles")

	config, err := configs.GetStableDiffusion(ctx, logCtx)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable get tti styles")
	}

	return c.JSON(http.StatusOK, config.Styles)
}
