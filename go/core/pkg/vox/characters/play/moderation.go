package play

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	fs "cloud.google.com/go/firestore"

	"disruptive/lib/common"
	"disruptive/lib/firestore"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/characters"
	"disruptive/pkg/vox/moderate"
	"disruptive/pkg/vox/profiles"
)

// GetSTSModeration get a moderation response for STS and save it to last_user_audio. Send the notification email.
func GetSTSModeration(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, characterVersion, audioID string) (*moderate.Response, error) {
	fid := slog.String("fid", "vox.moderate.GetSTSModeration")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	character, err := characters.GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("unable to get character", fid, "error", err)
		return &moderate.Response{}, err
	}

	session, err := characters.GetSession(ctx, logCtx, profile.ID, character.Character)
	if err != nil {
		logCtx.Error("unable to get session lastest document", fid, "error", err)
		return &moderate.Response{}, err
	}

	if len(session.LastUserAudio[audioID].Text) < 3 {
		return &moderate.Response{}, nil
	}

	m := moderate.Get(ctx, logCtx, session.LastUserAudio[audioID].Text, profileCharacter.Language)
	if m == nil {
		logCtx.Error("unable to generate moderation", fid)
		return &moderate.Response{}, errors.New("unable to generate moderation")
	}

	m.Triggered, m.Analysis.NotAgeAppropriate = validateModeration(profile, m)

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profile.ID, character.Character)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("unable to find collection", fid, "path", path)
		return &moderate.Response{}, nil
	}

	updates := []fs.Update{
		{Path: fmt.Sprintf("last_user_audio.%s.moderation", audioID), Value: m},
	}

	if _, err := collection.Doc("latest").Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to update latest document", fid, "error", err)
		return &moderate.Response{}, err
	}

	return m, nil
}

func validateModeration(profile *profiles.Document, m *moderate.Response) (bool, bool) {
	if profile.ResponseAge < min(m.Analysis.AssessmentAge, ratingsToAge[m.Analysis.MovieRating], ratingsToAge[m.Analysis.TVRating], ratingsToAge[m.Analysis.ESRBRating], m.Analysis.PEGIRating) {
		return true, true
	}

	return profile.Notifications.Moderations.Hate && m.Categories["hate"] ||
		profile.Notifications.Moderations.HateThreatening && m.Categories["hate/threatening"] ||
		profile.Notifications.Moderations.Harassment && m.Categories["harassment"] ||
		profile.Notifications.Moderations.HarassmentThreatening && m.Categories["harassment/threatening"] ||
		profile.Notifications.Moderations.Violence && m.Categories["violence"] ||
		profile.Notifications.Moderations.ViolenceGraphic && m.Categories["violence/graphic"] ||
		profile.Notifications.Moderations.Sexual && m.Categories["sexual"] ||
		profile.Notifications.Moderations.SexualMinors && m.Categories["sexual/minors"] ||
		profile.Notifications.Moderations.Selfharm && m.Categories["self-harm"] ||
		profile.Notifications.Moderations.SelfharmIntent && m.Categories["self-harm/intent"] ||
		profile.Notifications.Moderations.SelfharmInstructions && m.Categories["self-harm/instructions"] ||
		profile.Notifications.TextAnalysis.Toxic && m.Analysis.Toxic, false
}
