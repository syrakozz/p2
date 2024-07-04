package play

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"disruptive/lib/common"
	"disruptive/lib/configs"
	"disruptive/lib/firestore"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/characters"
	"disruptive/pkg/vox/profiles"

	fs "cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

// STSText is a structure for text.
type STSText struct {
	Text string `json:"text"`
}

// PostUserText creates a session with the given text.
func PostUserText(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, characterVersion string, text STSText, predefined bool) (characters.UserAudio, error) {
	fid := slog.String("fid", "vox.characters.play.PostUserText")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	s, err := characters.GetSession(ctx, logCtx, profile.ID, characterName)
	if err != nil {
		logCtx.Error("unable to get session", fid, "error", err)
		return characters.UserAudio{}, err
	}

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profile.ID, characterName)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("memory collection not found", fid)
		return characters.UserAudio{}, common.ErrNotFound{}
	}

	now := time.Now()
	audioID := uuid.New().String()

	userAudio := characters.UserAudio{
		AudioID:   audioID,
		Timestamp: now,
	}

	if text.Text == "" {
		userAudio.AudioID = "0"
		return userAudio, nil
	}

	if predefined {
		localize, err := configs.GetLocalization(ctx, logCtx, strings.Split(characterVersion, "_")[1], profileCharacter.Language)
		if err != nil {
			logCtx.Error("unable to get character localize configs", fid, "error", err)
			return userAudio, err
		}

		t, ok := localize.Predefined[text.Text]
		if !ok {
			logCtx.Error("unable to get predefined text from localize file", fid)
			return userAudio, common.ErrNotFound{Msg: "predefined text not found in localization"}
		}

		userAudio.Text = t
		userAudio.Predefined = true
		userAudio.Mode = text.Text
	} else {
		userAudio.Text = text.Text
		userAudio.Mode = profileCharacter.Mode
	}

	c, err := characters.GetCharacter(ctx, logCtx, characterVersion, profile.Characters[characterName].Language)
	if err != nil {
		logCtx.Error("unable to get character", fid, "error", err)
		return characters.UserAudio{}, err
	}

	tier := c.Modes[userAudio.Mode].Tier
	if tier == "" {
		tier = "tier-free"
	}

	if !account.DisableBank && tier != "tier-free" {
		balance, err := accounts.GetAvailableBalance(ctx, logCtx)
		if err != nil {
			logCtx.Error("unable to get account balance", fid, "error", err)
			return characters.UserAudio{}, err
		}

		if balance.TotalBalance < 1 {
			logCtx.Error("insufficient balance", fid, "error", err)
			return characters.UserAudio{}, common.ErrPaymentRequired{Msg: "insufficient balance", Src: "PostUserText"}
		}
	}

	if s.LastUserAudio == nil {
		s.Archive = now
		s.LastUserAudio = map[string]characters.UserAudio{}
	}
	s.LastUserAudio[audioID] = userAudio

	if s.StartEntry == 0 {
		if !userAudio.Predefined {
			s.StartEntry = 1
		}

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
