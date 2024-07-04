package play

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	fs "cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/google/uuid"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/deepgram"
	"disruptive/lib/firebase"
	"disruptive/lib/firestore"
	"disruptive/lib/gcp"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/characters"
	"disruptive/pkg/vox/profiles"
)

// PostUserAudio retrieves audio, creates a new session_id, and updates last_user_audio.
func PostUserAudio(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, characterVersion, format, version string, r io.Reader) (characters.UserAudio, error) {
	fid := slog.String("fid", "vox.characters.play.PostUserAudio")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	c, err := characters.GetCharacter(ctx, logCtx, characterVersion, profile.Characters[characterName].Language)
	if err != nil {
		logCtx.Error("unable to get character", fid, "error", err)
		return characters.UserAudio{}, err
	}

	tier := c.Modes[profileCharacter.Mode].Tier
	if tier == "" {
		tier = "tier-free"
	}

	if !account.DisableBank && tier != "tier-free" {
		balance, err := accounts.GetAvailableBalance(ctx, logCtx)
		if err != nil {
			logCtx.Error("unable to get account balance", fid, "error", err)
			return characters.UserAudio{}, err
		}

		switch tier {
		case "tier-conversation-1":
			if balance.Balance <= 20500 && balance.Balance > 20400 {
				emailLowBalance(ctx, logCtx, 20500)
			} else if balance.Balance <= 10500 && balance.Balance > 10400 {
				emailLowBalance(ctx, logCtx, 10500)
			}
		case "tier-fun-1":
			if balance.Balance <= 20500 && balance.Balance > 20350 {
				emailLowBalance(ctx, logCtx, 20500)
			} else if balance.Balance <= 10500 && balance.Balance > 10350 {
				emailLowBalance(ctx, logCtx, 10500)
			}
		case "tier-story-1":
			if balance.Balance <= 20500 && balance.Balance > 20300 {
				emailLowBalance(ctx, logCtx, 20500)
			} else if balance.Balance <= 10500 && balance.Balance > 10300 {
				emailLowBalance(ctx, logCtx, 10500)
			}
		}

		if balance.TotalBalance < 1 {
			logCtx.Error("insufficient balance", fid, "error", err)
			return characters.UserAudio{}, common.ErrPaymentRequired{Msg: "insufficient balance", Src: "PostUserAudio"}
		}
	}

	s, err := characters.GetSession(ctx, logCtx, profile.ID, characterName)
	if err != nil {
		logCtx.Error("unable to get session", fid, "error", err)
		return characters.UserAudio{}, err
	}

	fileExt, ok := deepgram.AudioFormatExtensions[format]
	if !ok {
		logCtx.Error("invalid audio format", fid, "format", format)
		return characters.UserAudio{}, common.ErrBadRequest{Msg: "invalid audio format"}
	}

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profile.ID, characterName)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("memory collection not found", fid)
		return characters.UserAudio{}, common.ErrNotFound{}
	}

	now := time.Now()
	audioID := uuid.New().String()

	if s.Archive.IsZero() {
		s.Archive = now
	}

	filename := fmt.Sprintf("%s.%s", audioID, fileExt)
	gcsPath := filepath.Join("accounts", account.ID, "profiles", profile.ID, "characters", characterName, "archives", s.Archive.Format(time.DateOnly), filename)

	// STT

	sttResponse, detectedLanguage, err := STT(ctx, logCtx, fileExt, profile.Characters[characterName].Language, version, gcsPath, r)
	if err != nil {
		logCtx.Error("unable to get speech-to-text response", fid, "error", err)
		return characters.UserAudio{}, err
	}

	userAudio := characters.UserAudio{Timestamp: now}
	if sttResponse.Text == "" {
		userAudio.AudioID = "0"
		return userAudio, nil
	}

	userAudio.AudioID = audioID
	userAudio.DetectedLanguage = detectedLanguage
	userAudio.Path = gcsPath
	userAudio.Text = sttResponse.Text
	userAudio.Mode = profile.Characters[characterName].Mode

	if s.LastUserAudio == nil {
		s.Archive = now
		s.LastUserAudio = map[string]characters.UserAudio{}
	}
	s.LastUserAudio[audioID] = userAudio

	if s.StartEntry == 0 {
		s.StartEntry = 1

		if _, err := collection.Doc("latest").Set(ctx, s); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to set latest document", fid, "error", err)
			return characters.UserAudio{}, err
		}
	} else {
		updates := []fs.Update{
			{Path: "last_user_audio." + audioID, Value: s.LastUserAudio[audioID]},
		}

		if _, err := collection.Doc("latest").Update(ctx, updates); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to update lastest document", fid, "error", err)
			return characters.UserAudio{}, err
		}
	}

	return userAudio, nil
}

// GetUserAudio retrieves user text from the session memory returning an io.Reader to the audio file in the format requested.
func GetUserAudio(ctx context.Context, logCtx *slog.Logger, profileID, characterVersion, format string, sessionID int) (io.ReadCloser, string, error) {
	fid := slog.String("fid", "vox.characters.play.GetUserAudio")

	e, err := characters.GetSessionEntryByID(ctx, logCtx, profileID, characterVersion, sessionID)
	if err != nil {
		logCtx.Error("unable to get session entry", fid, "error", err, "session_id", sessionID)
		return nil, "", err
	}

	fileExt, ok := common.AudioFormatExtensions[format]
	if !ok {
		logCtx.Error("invalid audio format", fid, "format", format)
		return nil, "", common.ErrBadRequest{Msg: "invalid audio format"}
	}

	if e.UserAudio[fileExt] == "" {
		logCtx.Error("user audio not found", fid, "format", format)
		return nil, "", common.ErrNotFound{Msg: "user audio not found"}
	}

	rc, cType, err := gcp.Storage.Download(ctx, firebase.GCSBucket, e.UserAudio[fileExt])
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			logCtx.Error("unable to download user audio file", fid, "error", err)
			return nil, "", common.ErrGone{Msg: "user audio gone"}
		}

		logCtx.Error("unable to download user audio file", fid, "error", err)
		return nil, "", err
	}

	return rc, cType, nil
}

// GetAssistantAudio runs a TTS. If the audio file exists in GCS it will return it. If not it will regenerate it.
func GetAssistantAudio(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, characterVersion, format, model, optimizingStreamLatency string, sessionID int, predefined bool) (io.ReadCloser, string, error) {
	fid := slog.String("fid", "vox.characters.play.GetUserAudio")

	rc, cType, err := TTS(ctx, logCtx, profile, characterVersion, format, model, optimizingStreamLatency, sessionID, predefined)
	if err != nil {
		logCtx.Error("unable to get text-to-speech response", fid, "error", err)
		return nil, "", err
	}

	return rc, cType, nil
}

const lowBalanceHTML = `
<!DOCTYPE html>
<html>
<body>
<h2>Low Vexels</h2>
<ul>
	<li><b>Email:</b> {{ .Email }}</li>
	<li><b>ID:</b> {{ .ID }}</li>
	<li><b>Project:</b> {{ .Project }}</li>
	<li><b>Balance:</b> {{ .Balance }}</li>
</ul>
</body>
</html>
`

type lowBalanceVars struct {
	Email   string
	ID      string
	Project string
	Balance int
}

func emailLowBalance(ctx context.Context, logCtx *slog.Logger, balance int) {
	fid := slog.String("fid", "vox.characters.play.emailLowBalance")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	logCtx.Info("low balance", fid, "email", account.Email, "id", account.ID)

	t, err := template.New("t").Parse(lowBalanceHTML)
	if err != nil {
		logCtx.Warn("unable to create low balance HTML template", fid, "error", err)
		return
	}

	var buf bytes.Buffer

	vars := lowBalanceVars{
		Email:   account.Email,
		ID:      account.ID,
		Project: config.VARS.FirebaseProject,
		Balance: balance,
	}

	if err := t.Execute(&buf, vars); err != nil {
		logCtx.Warn("unable to execute low balance HTML template", fid, "error", err)
		return
	}

	req := accounts.EmailRequest{
		From:    config.VARS.MailgunNotificationsFrom,
		To:      strings.Split(config.VARS.MailgunLowBalanceNotificationsTo, ","),
		Subject: "Low Vexels",
		HTML:    buf.String(),
	}

	if err := accounts.SendEmail(ctx, logCtx, req); err != nil {
		logCtx.Warn("unable to send mail", fid, "error", err)
		return
	}
}
