package characters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	fs "cloud.google.com/go/firestore"

	"github.com/pkoukk/tiktoken-go"
	"google.golang.org/api/iterator"

	"disruptive/lib/common"
	"disruptive/lib/firestore"
	"disruptive/lib/openai"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/profiles"
)

const systemPromptTemplate = `Create a concise summary of the following conversation between %s and %s by topic. 
Each topic should include the topic, 
a user summary of the questions asked by the user, 
a topic summary of the entire conversation,
and an analysis that gives a brief analysis of the conversation and its temperament.  
Combine similar prompts and responses into %d general topics.
Write the response in RFC8259 compliant JSON based on this JSON examples:

EXAMPLE:
[
    {
        "topic": "This is my Topic",
        "topic_summary": "This is the topic summary.",
        "user_summary": "This is a user summary.",
        "analysis": "This is an analysis of the conversation."
    }
]`

// SummaryResponseEntry contains moderation information from all services.
type SummaryResponseEntry struct {
	Analysis     string `firestore:"analysis,omitempty" json:"analysis,omitempty"`
	Topic        string `firestore:"topic,omitempty" json:"topic,omitempty"`
	TopicSummary string `firestore:"topic_summary,omitempty" json:"topic_summary,omitempty"`
	UserSummary  string `firestore:"user_summary,omitempty" json:"user_summary,omitempty"`
}

// ArchiveIndex contains a map of character archive entries
type ArchiveIndex map[string]ArchiveIndexEntry

// ArchiveIndexEntry contains a single character archive list entry.
type ArchiveIndexEntry struct {
	StartEntry int       `firestore:"start_entry" json:"start_entry"`
	StartTime  time.Time `firestore:"start_time" json:"start"`
	EndEntry   int       `firestore:"end_entry" json:"end_entry"`
	EndTime    time.Time `firestore:"end_time" json:"end"`
}

// ArchiveEntries is a map of archive entries.
type ArchiveEntries struct {
	Entries    map[string]SessionEntry `firestore:"entries" json:"entries"`
	StartEntry int                     `firestore:"start_entry" json:"start_entry"`
}

// GetArchiveByID retrieves a character archive by archive ID.
func GetArchiveByID(ctx context.Context, logCtx *slog.Logger, archiveID, profileID, characterVersion string) (SessionDocument, error) {
	fid := slog.String("fid", "vox.characters.GetArchiveByID")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Error("unable to get profile", fid, "error", err)
		return SessionDocument{}, err
	}

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	characterDoc, err := GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("character doc not found", fid, "error", err)
		return SessionDocument{}, err
	}

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profileID, characterDoc.Character)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("memory collection not found", fid)
		return SessionDocument{}, common.ErrNotFound{Msg: "collection not found"}
	}

	doc, err := collection.Doc(archiveID).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get character's archive document", fid, "error", err)
		return SessionDocument{}, err
	}

	d := SessionDocument{}

	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read archive data", fid, "error", err)
		return SessionDocument{}, err
	}

	return d, nil
}

// GetArchiveIndex returns a character's archive index.
func GetArchiveIndex(ctx context.Context, logCtx *slog.Logger, profileID, characterVersion string) (ArchiveIndex, error) {
	fid := slog.String("fid", "vox.characters.GetArchiveIndex")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Error("unable to get profile", fid, "error", err)
		return ArchiveIndex{}, err
	}

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	characterDoc, err := GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("character doc not found", fid, "error", err)
		return nil, err
	}

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profileID, characterDoc.Character)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Warn("memory collection not found", fid)
		return ArchiveIndex{}, common.ErrNotFound{Msg: "collection not found"}
	}

	doc, err := collection.Doc("index").Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		if !errors.Is(err, common.ErrNotFound{Err: err}) {
			logCtx.Error("unable to get character archive index", fid, "error", err)
			return nil, err
		}
	}

	a := ArchiveIndex{}

	if doc.Exists() {
		if err := doc.DataTo(&a); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to read archive data", fid, "error", err)
			return ArchiveIndex{}, err
		}
	}

	latest, err := collection.Doc("latest").Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get archive latest document", fid, "error", err)
		return ArchiveIndex{}, err
	}

	session := SessionDocument{}

	if err := latest.DataTo(&session); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read latest archive data", fid, "error", err)
		return ArchiveIndex{}, err
	}

	a["latest"] = ArchiveIndexEntry{
		StartEntry: session.StartEntry,
		StartTime:  session.Entries[fmt.Sprintf("%06d", session.StartEntry)].Timestamp,
		EndEntry:   session.StartEntry + len(session.Entries) - 1,
		EndTime:    session.Entries[fmt.Sprintf("%06d", session.StartEntry+len(session.Entries)-1)].Timestamp,
	}

	return a, nil
}

// GetArchiveEntriesByDateRange retrieves archive entries for a given date range.
func GetArchiveEntriesByDateRange(ctx context.Context, logCtx *slog.Logger, profileID, characterVersion string, startDate, endDate time.Time) (ArchiveEntries, error) {
	fid := slog.String("fid", "vox.characters.GetArchiveEntriesByDateRange")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Error("unable to get profile", fid, "error", err)
		return ArchiveEntries{}, err
	}

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	characterDoc, err := GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("character doc not found", fid, "error", err)
		return ArchiveEntries{}, err
	}

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/archives", account.ID, profileID, characterDoc.Character)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Warn("archives collection not found", fid)
		return ArchiveEntries{}, common.ErrNotFound{}
	}

	archiveIndex, err := GetArchiveIndex(ctx, logCtx, profileID, characterVersion)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get character's archive entry list", fid, "error", err)
		return ArchiveEntries{}, err
	}

	// determine which archives have entries that land within the specified date range
	archiveIDs := make([]string, 0, len(archiveIndex))
	var latestTime time.Time
	for aid, e := range archiveIndex {
		if (e.StartTime.Equal(startDate) || e.StartTime.After(startDate)) && (e.EndTime.Equal(endDate) || e.EndTime.Before(endDate)) {
			archiveIDs = append(archiveIDs, aid)
			if e.EndTime.After(latestTime) {
				latestTime = e.EndTime
			}
		}
	}

	if endDate.After(latestTime) {
		archiveIDs = append(archiveIDs, "latest")
	}

	if len(archiveIDs) < 1 {
		logCtx.Error("no archives for given date range", fid)
		return ArchiveEntries{}, common.ErrNoResults
	}

	archives := make([]SessionDocument, 0, len(archiveIDs))
	for _, aid := range archiveIDs {
		a, err := GetArchiveByID(ctx, logCtx, aid, profileID, characterVersion)
		if err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to get character's archive document", fid, "error", err)
			return ArchiveEntries{}, err
		}

		archives = append(archives, a)
	}

	if len(archives) < 1 {
		logCtx.Error("no archives found", fid)
		return ArchiveEntries{}, common.ErrNotFound{}
	}

	entries := ArchiveEntries{
		Entries: make(map[string]SessionEntry),
	}

	startEntry := maxNumberEntries
	for _, archive := range archives {
		for eid, entry := range archive.Entries {
			if (entry.Timestamp.Equal(startDate) || entry.Timestamp.After(startDate)) && (entry.Timestamp.Equal(endDate) || entry.Timestamp.Before(endDate)) {
				entries.Entries[eid] = entry
				if startEntry > eid {
					startEntry = eid
				}
			}
		}
	}

	if len(entries.Entries) < 1 {
		logCtx.Warn("no entries found within date range", fid)
		return ArchiveEntries{}, common.ErrNoResults
	}

	entries.StartEntry, err = strconv.Atoi(startEntry)
	if err != nil {
		logCtx.Error("unable to convert start entry string to int", fid, "error", err)
		return ArchiveEntries{}, err
	}

	return entries, nil
}

// GetArchiveSummaryByDateRange retrieves a summary of a character's archives within a specified date range.
// If it does not exist, create one, and save it to firebase.
func GetArchiveSummaryByDateRange(ctx context.Context, logCtx *slog.Logger, profileID, characterVersion string, startDate, endDate time.Time) ([]SummaryResponseEntry, error) {
	fid := slog.String("fid", "vox.characters.GetArchiveSummaryByDateRange")

	// check if summary already exists for given date range
	account := ctx.Value(common.AccountKey).(accounts.Document)

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Error("unable to get profile", fid, "error", err)
		return nil, err
	}

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	characterDoc, err := GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("character doc not found", fid, "error", err)
		return nil, err
	}

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/archives", account.ID, profileID, characterDoc.Character)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("archives collection not found", fid)
		return nil, common.ErrNotFound{}
	}

	doc, err := collection.Doc("date_range").Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		if !errors.Is(err, common.ErrNotFound{Err: err}) {
			logCtx.Error("unable to get character archives document", fid, "error", err)
			return nil, err
		}
	}

	summaryID := startDate.Truncate(time.Minute).Format(time.RFC3339) + " " + endDate.Add(time.Minute).Truncate(time.Minute).Format(time.RFC3339)
	summaryID = strings.ReplaceAll(summaryID, ":00Z", "")

	logCtx = logCtx.With("date_range", summaryID)

	a := map[string][]SummaryResponseEntry{}
	if doc.Exists() {
		if err := doc.DataTo(&a); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to read archive summary data", fid, "error", err)
			return nil, err
		}

		if s, ok := a[summaryID]; ok {
			return s, nil
		}
	}

	logCtx.Info("creating summary")

	// Archive summary does not exist for this date range, create it.

	archiveIndex, err := GetArchiveIndex(ctx, logCtx, profileID, characterVersion)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get character's archive entry list", fid, "error", err)
		return nil, err
	}

	// determine which archives have entries that land within the specified date range
	archiveIDs := make([]string, 0, len(archiveIndex))
	var latestTime time.Time
	for aid, e := range archiveIndex {
		if (e.StartTime.Equal(startDate) || e.StartTime.After(startDate)) && (e.EndTime.Equal(endDate) || e.EndTime.Before(endDate)) {
			archiveIDs = append(archiveIDs, aid)
			if e.EndTime.After(latestTime) {
				latestTime = e.EndTime
			}
		}
	}

	if endDate.After(latestTime) {
		archiveIDs = append(archiveIDs, "latest")
	}

	if len(archiveIDs) < 1 {
		logCtx.Error("no archives for given date range", fid)
		return nil, common.ErrNoResults
	}

	slices.Sort(archiveIDs)

	// get the actual archives needed for the summary
	archives := make([]SessionDocument, 0, len(archiveIDs))
	for _, aid := range archiveIDs {
		a, err := GetArchiveByID(ctx, logCtx, aid, profileID, characterVersion)
		if err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to get character's archive document", fid, "error", err)
			return nil, err
		}

		archives = append(archives, a)
	}

	if len(archives) < 1 {
		logCtx.Error("no archives for given date range", fid)
		return nil, common.ErrNoResults
	}

	// build the gpt prompt
	userPrompt := strings.Builder{}
	count := 0
	for _, doc := range archives {
		for _, e := range doc.Entries {
			if (e.Timestamp.Equal(startDate) || e.Timestamp.After(startDate)) && (e.Timestamp.Equal(endDate) || e.Timestamp.Before(endDate)) {
				userPrompt.WriteString(fmt.Sprintf("%s: %s\n", profile.Name, e.User))
				userPrompt.WriteString(fmt.Sprintf("%s: %s\n", characterDoc.Name, e.Assistant))
				count++
			}
		}
	}

	if count < 1 {
		logCtx.Warn("no entries for given date range", fid)
		return nil, common.ErrNoResults
	}

	numTopics := int(math.Ceil(1.75 * math.Log(float64(count)))) // dynamically increase number of topics based on size of conversation
	systemPrompt := fmt.Sprintf(systemPromptTemplate, profile.Name, characterDoc.Name, numTopics)

	// tokenize the prompt.
	tke, err := tiktoken.GetEncoding(TokenEncodingModel)
	if err != nil {
		logCtx.Error("unable to get tiktoken encoding", fid, "error", err)
		return nil, err
	}

	model := "gpt-3.5-turbo"
	numTokens := len(tke.Encode(systemPrompt+userPrompt.String(), nil, nil))
	if numTokens > MaxPromptTokens35Turbo {
		logCtx.Error("exceeded summary length", fid, "len", numTokens)
		return nil, common.ErrBadRequest{Msg: "Request exceeds max token length. Choose a smaller date range."}
	}

	// openai request and response
	chatReq := openai.ChatRequest{
		Model: model,
		Messages: []openai.ChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt.String()},
		},
	}

	chatRes, err := openai.PostChat(ctx, logCtx, chatReq)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logCtx.Error("timeout", fid, "error", err)
			return nil, err
		}
		logCtx.Error("unable to get openai chat response", "error", err)
		return nil, err
	}

	entries := []SummaryResponseEntry{}
	if err := json.Unmarshal([]byte(chatRes.Text), &entries); err != nil {
		logCtx.Error("unable to unmarshal response", fid, "error", err)
		return nil, err
	}

	docUpdate := map[string][]SummaryResponseEntry{
		summaryID: entries,
	}

	if _, err := collection.Doc("date_range").Set(ctx, docUpdate, fs.MergeAll); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to update create summaries by date", fid, "error", err)
		return nil, err
	}

	return entries, nil
}

// DeleteArchiveSummaryDateRange clears a characters's archive date range summaries.
func DeleteArchiveSummaryDateRange(ctx context.Context, logCtx *slog.Logger, profileID, characterVersion string) error {
	fid := slog.String("fid", "vox.characters.DeleteArchiveSummaryDateRange")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Error("unable to get profile", fid, "error", err)
		return err
	}

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	characterDoc, err := GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("character doc not found", fid, "error", err)
		return err
	}

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/archives", account.ID, profileID, characterDoc.Character)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Warn("archives collection not found", fid)
		return common.ErrNotFound{}
	}

	if _, err := collection.Doc("date_range").Delete(ctx); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to delete date range summaries", fid, "error", err)
		return err
	}

	return nil
}

// DeleteSessionMemory clears a characters's latest session memory, or all session memory.
func DeleteSessionMemory(ctx context.Context, logCtx *slog.Logger, profileID, characterVersion string) error {
	fid := slog.String("fid", "vox.characters.DeleteSessionMemory")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	profile, err := profiles.GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Error("unable to get profile", fid, "error", err)
		return err
	}

	characterName := strings.Split(characterVersion, "_")[0]
	profileCharacter := profile.Characters[characterName]

	characterDoc, err := GetCharacter(ctx, logCtx, characterVersion, profileCharacter.Language)
	if err != nil {
		logCtx.Error("character doc not found", fid, "error", err)
		return err
	}

	path := fmt.Sprintf("accounts/%s/profiles/%s/vox_sessions/%s/memory", account.ID, profileID, characterDoc.Character)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Warn("memory collection not found", fid)
		return common.ErrNotFound{}
	}

	bw := firestore.Client.BulkWriter(ctx)

	for {
		iter := collection.Documents(ctx)
		defer iter.Stop()

		numDeleted := 0

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				logCtx.Error("unable to delete session memory", fid, "error", err)
				bw.End()
				return err
			}
			bw.Delete(doc.Ref)
			numDeleted++
		}

		if numDeleted == 0 {
			bw.End()
			break
		}

		bw.Flush()
	}

	return nil
}
