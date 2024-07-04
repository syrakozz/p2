package play

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/pkg/vox/characters/play"
	"disruptive/pkg/vox/profiles"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetTTT is the REST API for getting text to text.
func GetTTT(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.GetTTT")

	profileID := c.Param("profile_id")
	characterVersion := c.Param("character_version")

	req := struct {
		Text string `query:"text"`
	}{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if req.Text == "" {
		return e.ErrBad(logCtx, fid, "text required")
	}

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		e.ErrBad(logCtx, fid, "invalid profile")
	}

	res, err := play.TTT(ctx, logCtx, &profile, characterVersion, "", req.Text, "")
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get text to text")
	}

	return c.JSON(http.StatusOK, res)
}
