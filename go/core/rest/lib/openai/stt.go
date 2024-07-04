package openai

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"disruptive/lib/openai"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// PostSTTFile is the REST API for OpenAI STT.
func PostSTTFile(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.lib.openai.PostSTTFile")

	extension := strings.ToLower(c.QueryParam("extension"))
	language := strings.ToLower(c.QueryParam("language"))

	if extension == "" {
		extension = "mp3"
	}

	if language == "" {
		language = "en-US"
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

	text, err := openai.PostTranscriptionsText(ctx, logCtx, f, extension, language)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to create deepgram transcription")
	}

	res := map[string]any{
		"extension": extension,
		"language":  language,
		"text":      text,
	}

	return c.JSON(http.StatusOK, res)
}

// PostSTTStream is the REST API for OpenAI STT.
func PostSTTStream(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.lib.openai.PostSTTStream")

	extension := strings.ToLower(c.QueryParam("extension"))
	language := strings.ToLower(c.QueryParam("language"))

	if extension == "" {
		extension = "mp3"
	}

	if language == "" {
		language = "en-US"
	}

	text, err := openai.PostTranscriptionsText(ctx, logCtx, c.Request().Body, extension, language)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to create deepgram transcription")
	}

	res := map[string]any{
		"extension": extension,
		"language":  language,
		"text":      text,
	}

	return c.JSON(http.StatusOK, res)
}
