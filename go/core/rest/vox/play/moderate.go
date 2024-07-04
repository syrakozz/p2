package play

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/pkg/vox/characters/play"
	"disruptive/pkg/vox/moderate"
	"disruptive/pkg/vox/profiles"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetSTSModeration is the REST API for getting moderation.
func GetSTSModeration(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.GetSTSModeration")

	profileID := c.Param("profile_id")
	characterVersion := c.Param("character_version")
	audioID := c.Param("audio_id")

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		e.ErrBad(logCtx, fid, "invalid profile")
	}

	if !profile.Moderate {
		return c.NoContent(http.StatusNoContent)
	}

	m, err := play.GetSTSModeration(ctx, logCtx, &profile, characterVersion, audioID)
	if err != nil {
		return e.ErrBad(logCtx, fid, "unable to get moderation")
	}

	return c.JSON(http.StatusOK, m)
}

// PostModerate is the REST API for moderation.
func PostModerate(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.PostModerate")

	req := struct {
		Language string `json:"language"`
		Text     string `json:"text"`
	}{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if req.Text == "" {
		return e.ErrBad(logCtx, fid, "missing text")
	}

	if req.Language == "" {
		req.Language = "en-US"
	}

	m := moderate.Get(ctx, logCtx, req.Text, req.Language)

	return c.JSON(http.StatusOK, m)
}
