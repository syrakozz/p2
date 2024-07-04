package play

import (
	"context"
	"io"
	"log/slog"
	"time"

	"disruptive/lib/common"
	"disruptive/lib/deepgram"
	"disruptive/lib/firebase"
	"disruptive/lib/gcp"
)

// STTResponse contains the response structure.
type STTResponse struct {
	AudioID string `json:"audio_id"`
	Text    string `json:"text"`
}

// STT retrieves audio and returns transcribed text.
func STT(ctx context.Context, logCtx *slog.Logger, extension, language, version, gcsPath string, r io.Reader) (STTResponse, string, error) {
	fid := slog.String("fid", "vox.characters.play.STT")
	t := time.Now()

	res := STTResponse{}

	var (
		contentType      string
		detectedLanguage string
		ok               bool
	)

	contentType, ok = common.AudioExtensionContentTypes[extension]
	if !ok {
		logCtx.Error("invalid audio format", fid, "extension", extension)
		return res, "", common.ErrBadRequest{Msg: "invalid audio format"}
	}

	rPipe, wPipe := io.Pipe()
	teeReader := io.TeeReader(r, wPipe)

	go func() {
		defer wPipe.Close()

		var (
			err  error
			text string
		)

		text, detectedLanguage, err = deepgram.PostTranscriptionsText(ctx, logCtx, teeReader, language, version)

		if err != nil {
			logCtx.Error("unable to create transcription", fid, "error", err)
			return
		}

		res.Text = text
	}()

	if err := gcp.Storage.Upload(ctx, rPipe, firebase.GCSBucket, gcsPath, contentType); err != nil {
		logCtx.Warn("unable to upload user audio", fid, "error", err)
		return res, "", err
	}

	logCtx.Info("duration", "duration", time.Since(t).Milliseconds(), "span", "stt")

	return res, detectedLanguage, nil
}
