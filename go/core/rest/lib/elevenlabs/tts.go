// Package elevenlabs interfaces directly with the elevenlabs APIs.
package elevenlabs

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/lib/elevenlabs"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// PostTTS is the REST API for Elevenlabs TTS.
func PostTTS(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.lib.elevenlabs.PostTTS")

	req := struct {
		Format   string `json:"format"`
		Language string `json:"language"`
		Text     string `json:"text"`
		Voice    string `json:"voice"`
	}{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if req.Text == "" {
		return e.ErrBad(logCtx, fid, "text required")
	}

	if req.Format == "" || req.Format == "mp3" || req.Format == "mp3_44100" {
		req.Format = "mp3_44100_128"
	} else if req.Format == "pcm" {
		req.Format = "pcm_16000"
	} else if req.Format == "opus" {
		req.Format = "opus_16000"
	}

	if _, ok := elevenlabs.AudioFormatExtensions[req.Format]; !ok {
		return e.ErrBad(logCtx, fid, "invalid audio format")
	}

	if req.Language == "" {
		req.Language = "en-US"
	}

	if req.Voice == "" {
		req.Voice = "2xl"
	}

	voice, ok := elevenlabs.Voices[req.Voice]
	if !ok {
		return e.ErrBad(logCtx, fid, "invalid voice")
	}

	elevenlabsReq := elevenlabs.Request{
		Format:            req.Format,
		Language:          req.Language,
		SimilarityBoost:   voice.SimilarityBoost,
		Stability:         voice.Stability,
		StyleExaggeration: voice.StyleExaggeration,
		Text:              req.Text,
		Voice:             voice.ID,
	}

	r, err := elevenlabs.TTSStream(ctx, logCtx, elevenlabsReq)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to create elevenlabs transcription")
	}
	contentType := elevenlabs.AudioFormatContentTypes[req.Format]

	return c.Stream(http.StatusOK, contentType, r)
}
