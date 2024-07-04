package play

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/elevenlabs"
	"disruptive/pkg/vox/characters/play"
	"disruptive/pkg/vox/profiles"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetAssistantAudio is the REST API for getting assistant audio.
func GetAssistantAudio(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.GetAssistantAudio")

	profileID := c.Param("profile_id")
	characterVersion := c.Param("character_version")
	sessionID := c.Param("session_id")
	format := strings.ToLower(c.QueryParam("format"))
	model := strings.ToLower(c.QueryParam("model"))
	optimizingStreamLatency := c.QueryParam("optimizing_stream_latency")
	predefined := strings.ToLower(c.QueryParam("predefined"))

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		e.ErrBad(logCtx, fid, "invalid profile")
	}

	if format == "" || format == "mp3" || format == "mp3_44100" {
		format = "mp3_44100_128"
	}

	if _, ok := common.AudioFormatExtensions[format]; !ok {
		return e.ErrBad(logCtx, fid, "invalid audio format")
	}

	if !slices.Contains(elevenlabs.OptimizingStreamLatency, optimizingStreamLatency) {
		return e.ErrBad(logCtx, fid, "invalid optimizing_stream_latency")
	}

	sid, err := strconv.Atoi(sessionID)
	if err != nil {
		return e.ErrBad(logCtx, fid, "invalid session_id")
	}

	r, cType, err := play.GetAssistantAudio(ctx, logCtx, &profile, characterVersion, format, model, optimizingStreamLatency, sid, predefined == "true")
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get audio")
	}
	defer r.Close()

	return c.Stream(http.StatusOK, cType, r)
}

// GetUserAudio is the REST API for getting user audio.
func GetUserAudio(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.GetUserAudio")

	profileID := c.Param("profile_id")
	characterVersion := c.Param("character_version")
	sessionID := c.Param("session_id")
	format := strings.ToLower(c.QueryParam("format"))

	if format == "" || format == "mp3" || format == "mp3_44100" {
		format = "mp3_44100_128"
	}

	if _, ok := common.AudioFormatExtensions[format]; !ok {
		return e.ErrBad(logCtx, fid, "invalid audio format")
	}

	sid, err := strconv.Atoi(sessionID)
	if err != nil {
		return e.ErrBad(logCtx, fid, "invalid session_id")
	}

	r, cType, err := play.GetUserAudio(ctx, logCtx, profileID, characterVersion, format, sid)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get audio")
	}
	defer r.Close()

	return c.Stream(http.StatusOK, cType, r)
}
