package play

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	fs "cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"

	"disruptive/lib/common"
	"disruptive/lib/configs"
	"disruptive/lib/elevenlabs"
	"disruptive/lib/firebase"
	"disruptive/lib/firestore"
	"disruptive/lib/gcp"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/characters"
	"disruptive/pkg/vox/profiles"
)

// TTS retrieves text from the session memory then TTS returning an io.Reader to the audio file.
func TTS(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, characterVersion, format, ttsModel, optimizingStreamLatency string, sessionID int, predefined bool) (io.ReadCloser, string, error) {
	fid := slog.String("fid", "vox.characters.play.TTS")

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	c, err := characters.GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("unable to get character", fid, "error", err)
		return nil, "", err
	}

	s, err := characters.GetSession(ctx, logCtx, profile.ID, c.Character)
	if err != nil {
		logCtx.Error("unable to get session lastest document", fid, "error", err)
		return nil, "", err
	}

	sessionIDStr := fmt.Sprintf("%06d", sessionID)

	var (
		e  characters.SessionEntry
		ok bool
	)
	if predefined {
		e, ok = s.PredefinedEntries[sessionIDStr]
		if !ok {
			logCtx.Error("text session entry not found", fid, "session_id", sessionID)
			return nil, "", common.ErrNotFound{Msg: "text session ID not found"}
		}
	} else {
		e, ok = s.Entries[sessionIDStr]
		if !ok {
			logCtx.Error("session entry not found", fid, "session_id", sessionID)
			return nil, "", common.ErrNotFound{Msg: "session ID not found"}
		}
	}

	if profileCharacter.Voice == "" {
		profileCharacter.Voice = "default"
	}

	if profileCharacter.Language == "" {
		profileCharacter.Language = "en-US"
	}

	fileExt, ok := elevenlabs.AudioFormatExtensions[format]
	if !ok {
		logCtx.Error("invalid audio format", fid, "format", format)
		return nil, "", common.ErrBadRequest{Msg: "invalid audio format"}
	}

	var (
		rc    io.ReadCloser
		cType string
		regen bool
	)

	if e.AssistantAudio[fileExt] != "" {
		rc, cType, err = gcp.Storage.Download(ctx, firebase.GCSBucket, e.AssistantAudio[fileExt])
		if err != nil {
			if !errors.Is(err, storage.ErrObjectNotExist) {
				logCtx.Error("unable to download assistant audio file", fid, "error", err)
				return nil, "", err
			}

			regen = true
		}
	}

	if e.AssistantAudio[fileExt] == "" || regen {
		v := elevenlabs.Voices[c.Voices[profileCharacter.Voice]]
		req := elevenlabs.Request{
			Format:                   format,
			Language:                 profileCharacter.Language,
			Model:                    ttsModel,
			OptimizeStreamingLatency: optimizingStreamLatency,
			Voice:                    v.ID,
			Text:                     e.Assistant,
			SimilarityBoost:          v.SimilarityBoost,
			Stability:                v.Stability,
			StyleExaggeration:        v.StyleExaggeration,
		}

		r, err := elevenlabs.TTSStream(ctx, logCtx, req)
		if err != nil {
			logCtx.Error("unable to get elevenlabs tts stream", fid, "error", err)
			return nil, "", err
		}

		contentType, ok := elevenlabs.AudioFormatContentTypes[format]
		if !ok {
			logCtx.Error("invalid audio format", fid, "format", format)
			return nil, "", common.ErrBadRequest{Msg: "invalid audio format"}
		}

		if format == "opus_16000" {
			return io.NopCloser(r), contentType, nil
		}

		rPipe, wPipe := io.Pipe()
		teeReader := io.TeeReader(r, wPipe)

		go func() {
			defer wPipe.Close()

			account := ctx.Value(common.AccountKey).(accounts.Document)

			archive := s.Archive.Format(time.DateOnly)

			var path string
			if predefined {
				path = sessionIDStr + "-predefined-assistant." + fileExt
			} else {
				path = sessionIDStr + "-assistant." + fileExt
			}

			gcsPath := filepath.Join("accounts", account.ID, "profiles", profile.ID, "characters", c.Character, "archives", archive, path)

			if err := gcp.Storage.Upload(ctx, teeReader, firebase.GCSBucket, gcsPath, contentType); err != nil {
				logCtx.Error("unable to store assistant audio file", fid, "error", err)
				return
			}

			fsPath := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profile.ID, c.Character)
			collection := firestore.Client.Collection(fsPath)
			if collection == nil {
				logCtx.Error("memory collection not fonud", fid, "error", err)
				return
			}

			var updates []fs.Update
			if predefined {
				updates = []fs.Update{
					{Path: fmt.Sprintf("predefined_entries.%s.assistant_audio.%s", sessionIDStr, fileExt), Value: gcsPath},
				}
			} else {
				updates = []fs.Update{
					{Path: fmt.Sprintf("entries.%s.assistant_audio.%s", sessionIDStr, fileExt), Value: gcsPath},
				}
			}

			if _, err := collection.Doc("latest").Update(ctx, updates); err != nil {
				err = common.ConvertGRPCError(err)
				logCtx.Warn("unable to update lastest document", fid, "error", err)
				return
			}
		}()

		return io.NopCloser(rPipe), contentType, nil
	}

	return rc, cType, nil
}

// TTSDontUnderstand converts dont_understand text to audio.
func TTSDontUnderstand(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, characterVersion, format string) (io.ReadCloser, string, error) {
	fid := slog.String("fid", "vox.characters.play.TTSDontUnderstand")
	t := time.Now()

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	c, err := characters.GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("unable to get character", fid, "error", err)
		return nil, "", err
	}

	if profileCharacter.Voice == "" {
		profileCharacter.Voice = "default"
	}

	if profileCharacter.Language == "" {
		profileCharacter.Language = "en-US"
	}

	localize, err := configs.GetLocalization(ctx, logCtx, strings.Split(characterVersion, "_")[1], profileCharacter.Language)
	if err != nil {
		logCtx.Error("unable to get character localize configs", fid, "error", err)
		return nil, "", err
	}

	v := elevenlabs.Voices[c.Voices[profileCharacter.Voice]]
	req := elevenlabs.Request{
		Format:            format,
		Language:          profileCharacter.Language,
		Voice:             v.ID,
		Text:              localize.Character["dont_understand"],
		SimilarityBoost:   v.SimilarityBoost,
		Stability:         v.Stability,
		StyleExaggeration: v.StyleExaggeration,
	}

	r, err := elevenlabs.TTSStream(ctx, logCtx, req)
	if err != nil {
		logCtx.Error("unable to get elevenlabs tts stream", fid, "error", err)
		return nil, "", err
	}

	contentType, ok := elevenlabs.AudioFormatContentTypes[format]
	if !ok {
		logCtx.Error("invalid audio format", fid, "format", format)
		return nil, "", common.ErrBadRequest{Msg: "invalid audio format"}
	}

	if format == "opus_16000" {
		return io.NopCloser(r), contentType, nil
	}

	logCtx.Info("duration", "duration", time.Since(t).Milliseconds(), "span", "tts")

	return io.NopCloser(r), contentType, nil
}

// TTSModerationResponse converts moderation_response text to audio.
func TTSModerationResponse(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, characterVersion, format string) (io.ReadCloser, string, error) {
	fid := slog.String("fid", "vox.characters.play.TTSModerationResponse")
	t := time.Now()

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	c, err := characters.GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("unable to get character", fid, "error", err)
		return nil, "", err
	}

	if profileCharacter.Voice == "" {
		profileCharacter.Voice = "default"
	}

	if profileCharacter.Language == "" {
		profileCharacter.Language = "en-US"
	}

	localize, err := configs.GetLocalization(ctx, logCtx, strings.Split(characterVersion, "_")[1], profileCharacter.Language)
	if err != nil {
		logCtx.Error("unable to get character localize configs", fid, "error", err)
		return nil, "", err
	}

	text := localize.Character["moderation_response_2"]
	if profile.ResponseAge > 12 {
		text = localize.Character["moderation_response_1"]
	}

	v := elevenlabs.Voices[c.Voices[profileCharacter.Voice]]
	req := elevenlabs.Request{
		Format:            format,
		Language:          profileCharacter.Language,
		Voice:             v.ID,
		Text:              text,
		SimilarityBoost:   v.SimilarityBoost,
		Stability:         v.Stability,
		StyleExaggeration: v.StyleExaggeration,
	}

	r, err := elevenlabs.TTSStream(ctx, logCtx, req)
	if err != nil {
		logCtx.Error("unable to get elevenlabs tts stream", fid, "error", err)
		return nil, "", err
	}

	contentType, ok := elevenlabs.AudioFormatContentTypes[format]
	if !ok {
		logCtx.Error("invalid audio format", fid, "format", format)
		return nil, "", common.ErrBadRequest{Msg: "invalid audio format"}
	}

	if format == "opus_16000" {
		return io.NopCloser(r), contentType, nil
	}

	logCtx.Info("duration", "duration", time.Since(t).Milliseconds(), "span", "tts")

	return io.NopCloser(r), contentType, nil
}
