package characters

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	fs "cloud.google.com/go/firestore"

	"disruptive/lib/common"
	"disruptive/lib/firestore"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/moderate"
	"disruptive/pkg/vox/profiles"
)

// SessionDocument contains session entries.
type SessionDocument struct {
	Entries           map[string]SessionEntry `firestore:"entries" json:"entries"`
	PredefinedEntries map[string]SessionEntry `firestore:"predefined_entries" json:"predefined_entries"`
	LastArchive       time.Time               `firestore:"last_archive" json:"last_archive"`
	Archive           time.Time               `firestore:"archive" json:"archive"`
	StartEntry        int                     `firestore:"start_entry" json:"start_entry"`
	LastUserAudio     map[string]UserAudio    `firestore:"last_user_audio" json:"last_user_audio"`
}

// SessionEntry contains a single user/assistant pair.
type SessionEntry struct {
	ID             int                `firestore:"id" json:"id,omitempty"`
	EndSequence    bool               `firestore:"end_sequence,omitempty" json:"end_sequence,omitempty"`
	User           string             `firestore:"user" json:"user,omitempty"`
	Assistant      string             `firestore:"assistant" json:"assistant,omitempty"`
	UserAudio      map[string]string  `firestore:"user_audio,omitempty" json:"user_audio,omitempty"`
	AssistantAudio map[string]string  `firestore:"assistant_audio" json:"assistant_audio,omitempty"`
	Mode           string             `firestore:"mode,omitempty" json:"mode,omitempty"`
	Moderation     *moderate.Response `firestore:"moderation,omitempty" json:"moderation,omitempty"`
	NotificationID string             `firestore:"notification_id,omitempty" json:"notification_id,omitempty"`
	Timestamp      time.Time          `firestore:"timestamp" json:"timestamp,omitempty"`
	TokensPrompt   int                `firestore:"tokens_prompt" json:"tokens_prompt,omitempty"`
	TokensResponse int                `firestore:"tokens_response" json:"tokens_response,omitempty"`
}

// UserAudio contains a user's audio info.
type UserAudio struct {
	AudioID          string             `firestore:"audio_id" json:"audio_id"`
	DetectedLanguage string             `firestore:"detected_language,omitempty" json:"detected_language,omitempty"`
	Predefined       bool               `firestore:"predefined,omitempty" json:"predefined,omitempty"`
	Mode             string             `firestore:"mode,omitempty" json:"mode,omitempty"`
	Moderation       *moderate.Response `firestore:"moderation,omitempty" json:"moderation,omitempty"`
	NotificationID   string             `firestore:"notification_id,omitempty" json:"notification_id,omitempty"`
	Path             string             `firestore:"path,omitempty" json:"path,omitempty"`
	SessionID        int                `firestore:"session_id" json:"session_id,omitempty"`
	Text             string             `firestore:"text" json:"text"`
	Timestamp        time.Time          `firestore:"timestamp" json:"timestamp"`
}

// GetSession returns the latest SessionDocument for an account/profile/character.
func GetSession(ctx context.Context, logCtx *slog.Logger, profileID string, character string) (SessionDocument, error) {
	fid := slog.String("fid", "vox.characters.GetSession")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profileID, character)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("memory collection not found", fid)
		return SessionDocument{}, common.ErrNotFound{}
	}

	s := SessionDocument{}

	doc, err := collection.Doc("latest").Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		if !errors.Is(err, common.ErrNotFound{}) {
			logCtx.Error("unable to get session document", fid, "error", err)
			return SessionDocument{}, err
		}

		logCtx.Info("creating latest doc", fid)

		_, err = collection.Doc("latest").Set(ctx, s)
		if err != nil {
			logCtx.Error("unable to create latest document", fid, "error", err)
			return SessionDocument{}, err
		}
	} else {
		if err := doc.DataTo(&s); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to read session data", fid, "error", err)
			return SessionDocument{}, err
		}
	}

	return s, nil
}

// GetSessionEntryByID returns the SessionEntry for an account/profile/character given a sessionID
func GetSessionEntryByID(ctx context.Context, logCtx *slog.Logger, profileID, characterVersion string, sessionID int) (SessionEntry, error) {
	fid := slog.String("fid", "vox.characters.GetSessionEntryByID")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profileID, characterVersion)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("memory collection not found", fid)
		return SessionEntry{}, common.ErrNotFound{}
	}

	archiveIndex, err := GetArchiveIndex(ctx, logCtx, profileID, characterVersion)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get character's archive entry list", fid, "error", err)
		return SessionEntry{}, err
	}

	var archiveID string
	for id, indexEntry := range archiveIndex {
		if indexEntry.StartEntry <= sessionID && indexEntry.EndEntry >= sessionID {
			archiveID = id
			break
		}
	}

	if archiveID == "" {
		logCtx.Warn("session not found", fid, "session_id", sessionID)
		return SessionEntry{}, common.ErrNotFound{}
	}

	s, err := GetArchiveByID(ctx, logCtx, archiveID, profileID, characterVersion)
	if err != nil {
		logCtx.Error("unable to get archive", fid, "error", err)
		return SessionEntry{}, err
	}

	return s.Entries[fmt.Sprintf("%06d", sessionID)], nil
}

// AddSessionEntry adds a new entry to the latest session document.
func AddSessionEntry(ctx context.Context, logCtx *slog.Logger, profileID, character, audioID string, session SessionDocument, sessionEntry SessionEntry) (int, error) {
	fid := slog.String("fid", "vox.characters.AddSessionEntry")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profileID, character)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("memory collection not found", fid)
		return 0, common.ErrNotFound{}
	}

	// Archive

	numEntries := len(session.Entries)
	sessionArchiveEntries := minSessionArchiveEntries + keepSessionEntries - 1

	if numEntries >= sessionArchiveEntries && sessionEntry.Timestamp.After(session.LastArchive.Add(24*time.Hour)) {
		logCtx = logCtx.With("mode", "archive")

		keepEntries := make(map[string]SessionEntry, keepSessionEntries)

		// Save archive
		sessionID := session.StartEntry + numEntries
		startEntry := session.StartEntry + numEntries - keepSessionEntries + 1
		if startEntry < 1 {
			startEntry = 1
		}

		for i := startEntry; i <= session.StartEntry+numEntries-1; i++ {
			keepStr := fmt.Sprintf("%06d", i)
			keepEntries[keepStr] = session.Entries[keepStr]
			delete(session.Entries, keepStr)
		}

		id := session.Archive.Format(time.DateOnly)
		session.LastUserAudio = nil
		session.PredefinedEntries = nil

		if _, err := collection.Doc(id).Set(ctx, session); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("set archive document", fid, "error", err)
			return 0, err
		}

		// Update latest
		sessionEntry.ID = sessionID
		keepEntries[fmt.Sprintf("%06d", sessionID)] = sessionEntry

		updates := []fs.Update{
			{Path: "entries", Value: keepEntries},
			{Path: "start_entry", Value: startEntry},
			{Path: "archive", Value: sessionEntry.Timestamp},
			{Path: "last_archive", Value: session.Archive},
		}

		if _, err := collection.Doc("latest").Update(ctx, updates); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to update lastest document", fid, "error", err)
			return 0, err
		}

		// Update character archive index
		logCtx = logCtx.With("mode", "update")

		path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profileID, character)
		collection := firestore.Client.Collection(path)
		if collection == nil {
			logCtx.Error("memory collection not found", fid)
			return 0, common.ErrNotFound{}
		}

		archive := ArchiveIndexEntry{
			StartEntry: session.StartEntry,
			StartTime:  session.Entries[fmt.Sprintf("%06d", session.StartEntry)].Timestamp,
			EndEntry:   session.StartEntry + len(session.Entries) - 1,
			EndTime:    session.Entries[fmt.Sprintf("%06d", session.StartEntry+len(session.Entries)-1)].Timestamp,
		}

		docUpdate := map[string]ArchiveIndexEntry{
			id: archive,
		}

		if _, err := collection.Doc("index").Set(ctx, docUpdate, fs.MergeAll); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to update archive index", fid, "error", err)
			return 0, err
		}

		return sessionID, nil
	}

	// No Archive, Update latest

	var (
		sessionID int
		updates   []fs.Update
	)

	if session.LastUserAudio[audioID].Predefined {
		sessionID = len(session.PredefinedEntries) + 1
		sessionEntry.ID = sessionID

		updates = []fs.Update{
			{Path: fmt.Sprintf("predefined_entries.%06d", sessionID), Value: sessionEntry},
		}
	} else {
		sessionID = session.StartEntry + len(session.Entries)
		sessionEntry.ID = sessionID

		updates = []fs.Update{
			{Path: fmt.Sprintf("entries.%06d", sessionID), Value: sessionEntry},
		}
	}

	if audioID != "" {
		updates = append(updates, fs.Update{Path: fmt.Sprintf("last_user_audio.%s.session_id", audioID), Value: sessionID})
	}

	if _, err := collection.Doc("latest").Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to update latest document", fid, "error", err)
		return 0, err
	}

	return sessionID, nil
}

// UpdateLastUserAudio copy the previous user audio at the top level to the correct session entry.
func UpdateLastUserAudio(ctx context.Context, logCtx *slog.Logger, profileID, character, audioID string, sessionID int, session SessionDocument) error {
	fid := slog.String("fid", "vox.characters.UpdateLastUserAudio")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	if _, ok := session.LastUserAudio[audioID]; !ok {
		logCtx.Warn("last user audio not found", fid, "audio_id", audioID)
		return common.ErrNotFound{}
	}

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profileID, character)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("unable to find collection", fid, "path", path)
		return common.ErrNotFound{}
	}

	sessionIDStr := fmt.Sprintf("%06d", sessionID)

	updates := []fs.Update{
		{Path: fmt.Sprintf("entries.%s.user", sessionIDStr), Value: session.LastUserAudio[audioID].Text},
	}

	if !session.LastUserAudio[audioID].Predefined {
		if session.LastUserAudio[audioID].Path != "" {
			ext := filepath.Ext(session.LastUserAudio[audioID].Path)
			if ext == "" {
				logCtx.Error("invalid extension", fid, "path", session.LastUserAudio)
				return common.ErrBadRequest{Msg: "invalid extension"}
			}

			updates = append(updates, fs.Update{Path: fmt.Sprintf("entries.%s.user_audio%s", sessionIDStr, ext), Value: session.LastUserAudio[audioID].Path})
		}

		updates = append(updates, fs.Update{Path: fmt.Sprintf("entries.%s.moderation", sessionIDStr), Value: session.LastUserAudio[audioID].Moderation})
	}

	for id, audio := range session.LastUserAudio {
		if time.Now().Sub(audio.Timestamp) > time.Duration(5*time.Minute) && id != audioID {
			updates = append(updates, fs.Update{Path: fmt.Sprintf("last_user_audio.%s", id), Value: fs.Delete})
		}
	}

	if _, err := collection.Doc("latest").Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to update latest document", fid, "error", err)
		return err
	}

	return nil
}

// EndSequence sets the latest session entry's end_sequence to true.
func EndSequence(ctx context.Context, logCtx *slog.Logger, profile *profiles.Document, characterVersion string) (SessionEntry, error) {
	fid := slog.String("fid", "vox.characters.EndSequence")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	character, err := GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("unable to get character", fid, "error", err)
		return SessionEntry{}, err
	}

	session, err := GetSession(ctx, logCtx, profile.ID, character.Character)
	if err != nil {
		logCtx.Error("unable to get session", fid, "error", err)
		return SessionEntry{}, err
	}

	if len(session.Entries) < 1 {
		logCtx.Info("no entries found", fid)
		return SessionEntry{}, nil
	}

	endSequenceID := len(session.Entries) + session.StartEntry - 1
	endSequenceIDStr := fmt.Sprintf("%06d", endSequenceID)

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profile.ID, character.Character)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("memory collection not found", fid)
		return SessionEntry{}, common.ErrNotFound{}
	}

	updates := []fs.Update{
		{Path: fmt.Sprintf("entries.%s.end_sequence", endSequenceIDStr), Value: true},
	}

	if _, err := collection.Doc("latest").Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to update latest document", fid, "error", err)
		return SessionEntry{}, err
	}

	session, err = GetSession(ctx, logCtx, profile.ID, character.Character)
	if err != nil {
		logCtx.Error("unable to get session", fid, "error", err)
		return SessionEntry{}, err
	}

	return session.Entries[endSequenceIDStr], nil
}
