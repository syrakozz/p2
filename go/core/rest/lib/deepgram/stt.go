// Package deepgram interfaces directly with the deepgram APIs.
package deepgram

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"disruptive/lib/deepgram"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// PostSTTFile is the REST API for Deepgram STT.
func PostSTTFile(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.lib.deepgram.PostSTTFile")

	mimeType := strings.ToLower(c.QueryParam("mimetype"))
	language := strings.ToLower(c.QueryParam("language"))
	version := strings.ToLower(c.QueryParam("version"))

	if mimeType == "" {
		mimeType = "audio/mpeg"
	}

	file, err := c.FormFile("file")
	if err != nil {
		return e.ErrBad(logCtx, fid, "invalid file")
	}

	f, err := file.Open()
	if err != nil {
		return e.ErrBad(logCtx, fid, "unable to open file")
	}
	defer f.Close()

	text, detectedLanguage, err := deepgram.PostTranscriptionsText(ctx, logCtx, f, language, version)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to create deepgram transcription")
	}

	res := map[string]any{
		"mimetype":          mimeType,
		"language":          language,
		"detected_language": detectedLanguage,
		"text":              text,
	}

	return c.JSON(http.StatusOK, res)
}

// PostSTTStream is the REST API for Deepgram STT.
func PostSTTStream(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.lib.deepgram.PostSTTStream")

	mimeType := strings.ToLower(c.QueryParam("mimetype"))
	language := strings.ToLower(c.QueryParam("language"))
	version := strings.ToLower(c.QueryParam("version"))

	if mimeType == "" {
		mimeType = "audio/mpeg"
	}

	text, detectedLanguage, err := deepgram.PostTranscriptionsText(ctx, logCtx, c.Request().Body, language, version)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to create deepgram transcription")
	}

	res := map[string]any{
		"mimetype":          mimeType,
		"language":          language,
		"detected_language": detectedLanguage,
		"text":              text,
	}

	return c.JSON(http.StatusOK, res)
}
