package play

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	fs "cloud.google.com/go/firestore"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/configs"
	"disruptive/lib/firestore"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/characters"
	"disruptive/pkg/vox/notifications"
	"disruptive/pkg/vox/profiles"
)

// CloseResponse is the response structure for sts/close.
type CloseResponse struct {
	ModerationEmailSent bool   `json:"moderation_email_sent"`
	NotificationID      string `json:"notification_id,omitempty"`
}

// GetSTSText retrieves last session entry.
func GetSTSText(ctx context.Context, logCtx *slog.Logger, profileID, characterVersion, audioID string) (characters.SessionEntry, error) {
	fid := slog.String("fid", "vox.characters.play.GetSTSText")

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Error("unable to get profile", fid, "error", err)
		return characters.SessionEntry{}, err
	}

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	c, err := characters.GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("unable to get character", fid, "error", err)
		return characters.SessionEntry{}, err
	}

	s, err := characters.GetSession(ctx, logCtx, profileID, c.Character)
	if err != nil {
		logCtx.Error("unable to get session lastest document", fid, "error", err)
		return characters.SessionEntry{}, err
	}

	sessionID := s.LastUserAudio[audioID].SessionID
	e, ok := s.Entries[fmt.Sprintf("%06d", sessionID)]
	if !ok {
		logCtx.Error("session entry not found", fid, "session_id", sessionID)
		return characters.SessionEntry{}, common.ErrNotFound{Msg: "session ID not found"}
	}

	return e, nil
}

// GetSTSAudio retrieves last_user_audio and returns audio.
func GetSTSAudio(ctx context.Context, logCtx *slog.Logger, profileID, characterVersion, format, tttModel, ttsModel, optimizingStreamLatency, audioID string) (io.ReadCloser, string, error) {
	fid := slog.String("fid", "vox.characters.play.GetSTSAudio")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Error("unable to get profile", fid, "error", err)
		return nil, "", err
	}

	if audioID == "0" {
		ttsResult, cType, err := processDontUnderstand(ctx, logCtx, &profile, characterVersion, format)
		if err != nil {
			logCtx.Error("unable to get text-to-speech response", fid, "error", err)
			return nil, "", err
		}

		return ttsResult, cType, nil
	}

	if audioID == "1" {
		ttsResult, cType, err := processModerationResponse(ctx, logCtx, &profile, characterVersion, format)
		if err != nil {
			logCtx.Error("unable to get text-to-speech response", fid, "error", err)
			return nil, "", err
		}

		return ttsResult, cType, nil
	}

	characterName := strings.Split(characterVersion, "_")[0]

	s, err := characters.GetSession(ctx, logCtx, profileID, characterName)
	if err != nil {
		logCtx.Error("unable to get session", fid, "error", err)
		return nil, "", err
	}

	sessionID := s.LastUserAudio[audioID].SessionID
	if profile.Characters[characterName].Language == "" {
		characterPref := profile.Characters[characterName]
		characterPref.Language = s.LastUserAudio[audioID].DetectedLanguage
		profile.Characters[characterName] = characterPref
	}

	var (
		tttResult *Result
	)

	if sessionID == 0 {
		tttResult, err = TTT(ctx, logCtx, &profile, characterVersion, tttModel, s.LastUserAudio[audioID].Text, audioID)
		if err != nil {
			logCtx.Error("unable to get text-to-text response", fid, "error", err)
			return nil, "", err
		}

		sessionID = tttResult.SessionID
	}

	ttsReader, cType, err := TTS(ctx, logCtx, &profile, characterVersion, format, ttsModel, optimizingStreamLatency, sessionID, s.LastUserAudio[audioID].Predefined)
	if err != nil {
		logCtx.Error("unable to get text-to-speech response", fid, "error", err)
		return nil, "", err
	}

	if account.DisableBank {
		return ttsReader, cType, nil
	}

	c, err := characters.GetCharacter(ctx, logCtx, characterVersion, profile.Characters[characterName].Language)
	if err != nil {
		logCtx.Error("unable to get character", fid, "error", err)
		return nil, "", err
	}

	tier := c.Modes[s.LastUserAudio[audioID].Mode].Tier
	if tier == "" {
		tier = "tier-free"
	}

	_, err = accounts.ChargeBank(ctx, logCtx, characterVersion, tier)
	if err != nil {
		logCtx.Warn("unable to charge account", fid, "error", err)
		return ttsReader, cType, err
	}

	return ttsReader, cType, nil
}

// GetSTSDontUnderstandText retrieves dont understand text.
func GetSTSDontUnderstandText(ctx context.Context, logCtx *slog.Logger, profileID, characterVersion string) (characters.SessionEntry, error) {
	fid := slog.String("fid", "vox.characters.play.GetSTSDontUnderstandText")

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Error("unable to get profile", fid, "error", err)
		return characters.SessionEntry{}, err
	}

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	localize, err := configs.GetLocalization(ctx, logCtx, strings.Split(characterVersion, "_")[1], profileCharacter.Language)
	if err != nil {
		logCtx.Error("unable to get character localize configs", fid, "error", err)
		return characters.SessionEntry{}, err
	}

	return characters.SessionEntry{Assistant: localize.Character["dont_understand"]}, nil
}

// GetSTSModerationResponseText retrieves predefined moderation response.
func GetSTSModerationResponseText(ctx context.Context, logCtx *slog.Logger, profileID, characterVersion string) (characters.SessionEntry, error) {
	fid := slog.String("fid", "vox.characters.play.GetSTSModerationResponseText")

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Error("unable to get profile", fid, "error", err)
		return characters.SessionEntry{}, err
	}

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	localize, err := configs.GetLocalization(ctx, logCtx, strings.Split(characterVersion, "_")[1], profileCharacter.Language)
	if err != nil {
		logCtx.Error("unable to get character localize configs", fid, "error", err)
		return characters.SessionEntry{}, err
	}

	if profile.ResponseAge > 12 {
		return characters.SessionEntry{Assistant: localize.Character["moderation_response_1"]}, nil
	}

	return characters.SessionEntry{Assistant: localize.Character["moderation_response_2"]}, nil
}

// CloseSTS finalizes the STS process. Updates the session entry with last_user_audio values, and sends moderation email if necessary.
func CloseSTS(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, characterVersion, audioID string) (CloseResponse, error) {
	fid := slog.String("fid", "vox.characters.play.CloseSTS")

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	character, err := characters.GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("unable to get character", fid, "error", err)
		return CloseResponse{}, err
	}

	session, err := characters.GetSession(ctx, logCtx, profile.ID, character.Character)
	if err != nil {
		logCtx.Error("unable to get session", fid, "error", err)
		return CloseResponse{}, err
	}

	if session.LastUserAudio[audioID].Predefined {
		return CloseResponse{}, nil
	}

	sessionID := session.LastUserAudio[audioID].SessionID
	if sessionID == 0 && session.LastUserAudio[audioID].Moderation == nil {
		return CloseResponse{}, nil
	}
	sessionIDStr := fmt.Sprintf("%06d", sessionID)

	language := profileCharacter.Language
	if language == "" && audioID != "" && !session.LastUserAudio[audioID].Predefined {
		language = session.LastUserAudio[audioID].DetectedLanguage
	}
	logCtx.Info("stt language", "region", config.VARS.STTRegion, "language", language)

	localize, err := configs.GetLocalization(ctx, logCtx, strings.Split(characterVersion, "_")[1], language)
	if err != nil {
		logCtx.Error("unable to get character localize configs", fid, "error", err)
		return CloseResponse{}, err
	}

	if audioID != "" && sessionID != 0 {
		if err := characters.UpdateLastUserAudio(ctx, logCtx, profile.ID, character.Character, audioID, sessionID, session); err != nil {
			logCtx.Error("unable to update last user audio", fid, "error", err)
			return CloseResponse{}, err
		}
	}

	session, err = characters.GetSession(ctx, logCtx, profile.ID, character.Character)
	if err != nil {
		logCtx.Error("unable to get session", fid, "error", err)
		return CloseResponse{}, err
	}

	resp := CloseResponse{}

	if profile.Moderate && session.LastUserAudio[audioID].Moderation != nil && session.LastUserAudio[audioID].Moderation.Triggered {
		account := ctx.Value(common.AccountKey).(accounts.Document)

		path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profile.ID, character.Character)
		collection := firestore.Client.Collection(path)
		if collection == nil {
			logCtx.Error("memory collection not found", fid)
			return CloseResponse{}, common.ErrNotFound{}
		}

		modText := localize.Character["moderation_response_2"]
		if profile.ResponseAge > 12 {
			modText = localize.Character["moderation_response_1"]
		}

		m := session.Entries[sessionIDStr]
		m.Assistant = modText
		session.Entries[sessionIDStr] = m

		req := notifications.ModerationValue{
			Profile: notifications.ModerationProfileValue{
				ID:   profile.ID,
				Name: profile.Name,
			},
			Character: notifications.ModerationCharacterValue{
				Name: characterName,
			},
			Session: &notifications.ModerationSessionValue{
				Archive:     session.Archive,
				Entry:       session.Entries[sessionIDStr],
				EntryNumber: sessionID,
			},
		}

		doc, err := notifications.PostModeration(ctx, logCtx, req)
		if err != nil {
			logCtx.Error("unable to create moderation notification", fid, "error", err)
			return CloseResponse{}, err
		}

		if sessionID != 0 {
			updates := []fs.Update{
				{Path: fmt.Sprintf("entries.%s.notification_id", sessionIDStr), Value: doc.ID},
				{Path: fmt.Sprintf("entries.%s.assistant", sessionIDStr), Value: modText},
				{Path: fmt.Sprintf("last_user_audio.%s.notification_id", audioID), Value: doc.ID},
			}

			if _, err := collection.Doc("latest").Update(ctx, updates); err != nil {
				err = common.ConvertGRPCError(err)
				logCtx.Error("unable to update latest document", fid, "error", err)
				return CloseResponse{}, err
			}
		}

		resp.NotificationID = doc.ID

		if len(profile.Notifications.Emails) < 1 {
			return resp, nil
		}

		lastUserAudio := session.LastUserAudio[audioID]
		if err := notifications.SendModerationEmail(ctx, logCtx, profile, character.Character, &lastUserAudio, sessionID, &localize); err != nil {
			logCtx.Warn("unable to create email html", fid, "error", err)
		}

		resp.ModerationEmailSent = true
	}

	return resp, nil
}

func processDontUnderstand(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, characterVersion, format string) (io.ReadCloser, string, error) {
	fid := slog.String("fid", "vox.characters.play.processDontUnderstand")

	ttsReader, cType, err := TTSDontUnderstand(ctx, logCtx, profile, characterVersion, format)
	if err != nil {
		logCtx.Error("unable to get text-to-speech don't understand response", fid, "error", err)
		return nil, "", err
	}

	return ttsReader, cType, nil
}

func processModerationResponse(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, characterVersion, format string) (io.ReadCloser, string, error) {
	fid := slog.String("fid", "vox.characters.play.processModerationResponse")

	ttsReader, cType, err := TTSModerationResponse(ctx, logCtx, profile, characterVersion, format)
	if err != nil {
		logCtx.Error("unable to get text-to-speech moderation response", fid, "error", err)
		return nil, "", err
	}

	return ttsReader, cType, nil
}
