package play

import (
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/elevenlabs"
	"disruptive/pkg/vox/characters"
	"disruptive/pkg/vox/characters/play"
	"disruptive/pkg/vox/profiles"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

var (
	websocketUpgrader = websocket.Upgrader{CheckOrigin: checkOrigin}
)

func checkOrigin(_ *http.Request) bool {
	return true
}

// PostSTSAudioFile is the REST API to upload user audio from a file.
func PostSTSAudioFile(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.PostSTSAudioFile")

	profileID := c.Param("profile_id")
	characterVersion := c.Param("character_version")
	format := strings.ToLower(c.QueryParam("format"))
	version := strings.ToLower(c.QueryParam("version"))

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		e.ErrBad(logCtx, fid, "invalid profile")
	}

	if format == "" {
		format = "mp3"
	}

	if _, ok := common.AudioExtensionContentTypes[format]; !ok {
		return e.ErrBad(logCtx, fid, "invalid audio format")
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

	res, err := play.PostUserAudio(ctx, logCtx, &profile, characterVersion, format, version, f)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process user audio")
	}

	return c.JSON(http.StatusCreated, res)
}

// PostSTSAudioStream is the REST API to upload user audio from a stream.
func PostSTSAudioStream(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.PostSTSAudioStream")

	profileID := c.Param("profile_id")
	characterVersion := c.Param("character_version")
	format := strings.ToLower(c.QueryParam("format"))
	version := strings.ToLower(c.QueryParam("version"))

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		e.ErrBad(logCtx, fid, "invalid profile")
	}

	if format == "" {
		format = "mp3"
	}

	if _, ok := common.AudioExtensionContentTypes[format]; !ok {
		return e.ErrBad(logCtx, fid, "invalid audio format")
	}

	res, err := play.PostUserAudio(ctx, logCtx, &profile, characterVersion, format, version, c.Request().Body)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process user audio")
	}

	return c.JSON(http.StatusCreated, res)
}

// PostSTSTextPredefined is the REST API to use predefined text.
func PostSTSTextPredefined(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.PostSTSTextPredefined")

	profileID := c.Param("profile_id")
	characterVersion := c.Param("character_version")

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		e.ErrBad(logCtx, fid, "invalid profile")
	}

	t := play.STSText{}

	if err := c.Bind(&t); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if t.Text == "" {
		return e.ErrBad(logCtx, fid, "text required")
	}

	res, err := play.PostUserText(ctx, logCtx, &profile, characterVersion, t, true)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process user text")
	}

	return c.JSON(http.StatusCreated, res)
}

// PostSTSText is the REST API to upload user text.
func PostSTSText(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.PostSTSText")

	profileID := c.Param("profile_id")
	characterVersion := c.Param("character_version")

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		e.ErrBad(logCtx, fid, "invalid profile")
	}

	t := play.STSText{}

	if err := c.Bind(&t); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if t.Text == "" {
		return e.ErrBad(logCtx, fid, "text required")
	}

	res, err := play.PostUserText(ctx, logCtx, &profile, characterVersion, t, false)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process user text")
	}

	return c.JSON(http.StatusCreated, res)
}

// GetSTSAudio is the REST API for speech-to-speech.
func GetSTSAudio(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.GetSTSAudio")

	profileID := c.Param("profile_id")
	audioID := c.Param("audio_id")
	characterVersion := c.Param("character_version")
	format := strings.ToLower(c.QueryParam("format"))
	ttsModel := strings.ToLower(c.QueryParam("tts_model"))
	tttModel := strings.ToLower(c.QueryParam("ttt_model"))
	optimizingStreamLatency := c.QueryParam("optimizing_stream_latency")

	if format == "" || format == "mp3" || format == "mp3_44100" {
		format = "mp3_44100_128"
	} else if format == "pcm" {
		format = "pcm_16000"
	} else if format == "opus" {
		format = "opus_16000"
	}

	if _, ok := elevenlabs.AudioFormatExtensions[format]; !ok {
		return e.ErrBad(logCtx, fid, "invalid audio format")
	}

	if !slices.Contains(elevenlabs.OptimizingStreamLatency, optimizingStreamLatency) {
		return e.ErrBad(logCtx, fid, "invalid optimizing_stream_latency")
	}

	t := time.Now()
	defer func() {
		logCtx.Info("duration", "duration", time.Since(t).Milliseconds(), "span", "tts")
	}()

	r, cType, err := play.GetSTSAudio(ctx, logCtx, profileID, characterVersion, format, tttModel, ttsModel, optimizingStreamLatency, audioID)
	if err != nil {
		if errors.Is(err, common.ErrNoResults) {
			return c.NoContent(http.StatusNoContent)
		}

		return e.Err(logCtx, err, fid, "unable to process speech-to-speech")
	}
	defer r.Close()

	return c.Stream(http.StatusOK, cType, r)
}

// GetSTSAudioWS is the REST API for speech-to-speech.
func GetSTSAudioWS(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.GetSTSAudioWS")

	ws, err := websocketUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logCtx.Error("unable to upgrade to websocket", fid, "error", err)
		return err
	}
	defer ws.Close()

	profileID := c.Param("profile_id")
	audioID := c.Param("audio_id")
	characterVersion := c.Param("character_version")
	format := strings.ToLower(c.QueryParam("format"))
	ttsModel := strings.ToLower(c.QueryParam("tts_model"))
	tttModel := strings.ToLower(c.QueryParam("ttt_model"))
	optimizingStreamLatency := c.QueryParam("optimizing_stream_latency")

	if format == "" || format == "mp3" || format == "mp3_44100" {
		format = "mp3_44100_128"
	} else if format == "pcm" {
		format = "pcm_16000"
	} else if format == "opus" {
		format = "opus_16000"
	}

	if _, ok := elevenlabs.AudioFormatExtensions[format]; !ok {
		logCtx.Error("invalid audio format", fid)
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, "invalid audio format"))
		return nil
	}

	if !slices.Contains(elevenlabs.OptimizingStreamLatency, optimizingStreamLatency) {
		return e.ErrBad(logCtx, fid, "invalid optimizing_stream_latency")
	}

	t := time.Now()
	defer func() {
		logCtx.Info("duration", "duration", time.Since(t).Milliseconds(), "span", "tts")
	}()

	r, _, err := play.GetSTSAudio(ctx, logCtx, profileID, characterVersion, format, tttModel, ttsModel, optimizingStreamLatency, audioID)
	if err != nil {
		if errors.Is(err, common.ErrNoResults) {
			ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return nil
		}
		logCtx.Error("unable to process speech-to-speech", fid, "error", err)
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, "unable to process speech-to-speech"))
		return nil
	}
	defer r.Close()

	// Stream audio to WebSocket
	buffer := make([]byte, 4096)
	for {
		n, err := r.Read(buffer)
		if n > 0 {
			if err := ws.WriteMessage(websocket.BinaryMessage, buffer[:n]); err != nil {
				logCtx.Error("unable to send audio data", fid, "error", err)
				ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, "unable to send audio data"))
				return nil
			}
		}

		if err == io.EOF {
			ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			break
		}

		if err != nil {
			ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, err.Error()))
			return nil
		}
	}

	return nil
}

// GetSTSText is the REST API for getting text entry.
func GetSTSText(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.GetSTSText")

	profileID := c.Param("profile_id")
	audioID := c.Param("audio_id")
	characterVersion := c.Param("character_version")

	var (
		res characters.SessionEntry
		err error
	)

	switch audioID {
	case "0":
		res, err = play.GetSTSDontUnderstandText(ctx, logCtx, profileID, characterVersion)
		if err != nil {
			return e.Err(logCtx, err, fid, "unable to get dont understand assistant text")
		}
	case "1":
		res, err = play.GetSTSModerationResponseText(ctx, logCtx, profileID, characterVersion)
		if err != nil {
			return e.Err(logCtx, err, fid, "unable to get moderation response assistant text")
		}
	default:
		res, err = play.GetSTSText(ctx, logCtx, profileID, characterVersion, audioID)
		if err != nil {
			return e.Err(logCtx, err, fid, "unable to get assistant text")
		}
	}

	return c.JSON(http.StatusOK, res)
}

// PostSTSClose is the REST API to finalize an STS.
func PostSTSClose(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.PostSTSClose")

	profileID := c.Param("profile_id")
	characterVersion := c.Param("character_version")
	audioID := c.Param("audio_id")

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		return e.ErrBad(logCtx, fid, "invalid profile")
	}

	resp, err := play.CloseSTS(ctx, logCtx, &profile, characterVersion, audioID)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process user audio")
	}

	return c.JSON(http.StatusOK, resp)
}

// PatchSTSEndSequence is the REST API to end a session memory sequence
func PatchSTSEndSequence(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.play.PatchSTSEndSequence")

	profileID := c.Param("profile_id")
	characterVersion := c.Param("character_version")

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		return e.ErrBad(logCtx, fid, "invalid profile")
	}

	resp, err := characters.EndSequence(ctx, logCtx, &profile, characterVersion)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process end sequence")
	}

	return c.JSON(http.StatusOK, resp)
}
