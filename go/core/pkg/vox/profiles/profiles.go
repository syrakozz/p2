package profiles

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	fs "cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"

	"disruptive/lib/common"
	"disruptive/lib/firestore"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/moderate"
)

// Document contains a firestore profile document.
type Document struct {
	ID                   string                          `firestore:"id" json:"id,omitempty"`
	AddQuestionFrequency int                             `firestore:"add_question_frequency" json:"add_question_frequency"`
	Characters           map[string]CharacterPreferences `firestore:"characters" json:"characters,omitempty"`
	CreatedDate          time.Time                       `firestore:"created_date" json:"created_date"`
	DontSay              []string                        `firestore:"dont_say" json:"dont_say,omitempty"`
	Inactive             bool                            `firestore:"inactive" json:"inactive,omitempty"`
	Interests            []string                        `firestore:"interests" json:"interests,omitempty"`
	Moderate             bool                            `firestore:"moderate" json:"moderate,omitempty"`
	ModifiedDate         time.Time                       `firestore:"modified_date" json:"modified_date"`
	Name                 string                          `firestore:"name" json:"name"`
	Notifications        Notifications                   `firestore:"notifications" json:"notifications,omitempty"`
	Picture              string                          `firestore:"picture" json:"picture"`
	Preferences          map[string]any                  `firestore:"preferences" json:"preferences"`
	ReplaceWords         map[string][]string             `firestore:"replace_words" json:"replace_words"`
	ResponseAge          int                             `firestore:"response_age" json:"response_age"`
	SelectedCharacter    string                          `firestore:"selected_character" json:"selected_character"`
	TopicsDiscourage     []string                        `firestore:"topics_discourage" json:"topics_discourage"`
	TopicsEncourage      []string                        `firestore:"topics_encourage" json:"topics_encourage"`
}

// PatchDocument contains a firestore profile document.
type PatchDocument struct {
	AddQuestionFrequency *int                  `json:"add_question_frequency"`
	DontSay              *[]string             `json:"dont_say"`
	Inactive             *bool                 `json:"inactive"`
	Interests            *[]string             `json:"interests"`
	Moderate             *bool                 `json:"moderate"`
	Name                 *string               `json:"name"`
	Notifications        *Notifications        `json:"notifications"`
	ReplaceWords         *map[string]*[]string `json:"replace_words"`
	ResponseAge          *int                  `json:"response_age"`
	SelectedCharacter    *string               `json:"selected_character"`
	TopicsDiscourage     *[]string             `json:"topics_discourage"`
	TopicsEncourage      *[]string             `json:"topics_encourage"`
}

// CharacterPreferences contains a firestore profile character document.
type CharacterPreferences struct {
	ImageStyle string `firestore:"image_style" json:"image_style"`
	Language   string `firestore:"language" json:"language"`
	Mode       string `firestore:"mode" json:"mode"`
	Voice      string `firestore:"voice" json:"voice"`
}

// PatchCharacterPreferences contains a firestore profile character document.
// TODO: remove these patch values from PatchDocument and Patch
type PatchCharacterPreferences struct {
	ImageStyle *string `json:"image_style"`
	Language   *string `json:"language"`
	Mode       *string `json:"mode"`
	Voice      *string `json:"voice"`
}

// Notifications contains notification indicators.
type Notifications struct {
	Emails      []string `firestore:"emails,omitempty" json:"emails,omitempty"`
	Moderations struct {
		Hate                  bool `firestore:"hate" json:"hate,omitempty"`
		HateThreatening       bool `firestore:"hate_threatening" json:"hate_threatening,omitempty"`
		Harassment            bool `firestore:"harassment" json:"harassment,omitempty"`
		HarassmentThreatening bool `firestore:"harassment_threatening" json:"harassment_threatening,omitempty"`
		Selfharm              bool `firestore:"selfharm" json:"selfharm,omitempty"`
		SelfharmIntent        bool `firestore:"selfharm_intent" json:"selfharm_intent,omitempty"`
		SelfharmInstructions  bool `firestore:"selfharm_instructions" json:"selfharm_instructions,omitempty"`
		Sexual                bool `firestore:"sexual" json:"sexual,omitempty"`
		SexualMinors          bool `firestore:"sexual_minors" json:"sexual_minors,omitempty"`
		Violence              bool `firestore:"violence" json:"violence,omitempty"`
		ViolenceGraphic       bool `firestore:"violence_graphic" json:"violence_graphic,omitempty"`
	} `firestore:"moderations" json:"moderations"`
	TextAnalysis struct {
		NotAgeAppropriate bool `firestore:"not_age_appropriate" json:"not_age_appropriate,omitempty"`
		Toxic             bool `firestore:"toxic" json:"toxic,omitempty"`
	} `firestore:"text_analysis" json:"text_analysis"`
}

// Create creates a new firestore profile document.
func Create(ctx context.Context, logCtx *slog.Logger, document Document) (Document, error) {
	fid := slog.String("fid", "vox.profiles.Create")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	res := moderate.Get(ctx, logCtx, document.Name, "en-US")

	for _, v := range res.Categories {
		if v {
			return Document{}, common.ErrBadRequest{Src: "profile", Msg: "invalid profile name"}
		}
	}

	if res.Analysis.Toxic {
		return Document{}, common.ErrBadRequest{Src: "profile", Msg: "invalid profile name"}
	}

	path := fmt.Sprintf("accounts/%s/profiles", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("profiles collection not found", fid)
		return Document{}, common.ErrNotFound{}
	}

	uuid := uuid.New().String()
	document.ID = uuid
	document.AddQuestionFrequency = 60
	document.CreatedDate = time.Now()

	if _, err := collection.Doc(uuid).Create(ctx, document); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to create accounts document", fid, "error", err)
		return Document{}, err
	}

	return document, nil
}

// Get returns all profiles
func Get(ctx context.Context, logCtx *slog.Logger, all, inactive bool) ([]Document, error) {
	fid := slog.String("fid", "vox.profiles.Get")

	user := ctx.Value(common.AccountKey).(accounts.Document)

	docs := []Document{}

	path := fmt.Sprintf("accounts/%s/profiles", user.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("profiles collection not found", fid)
		return nil, common.ErrNotFound{}
	}

	var iter *fs.DocumentIterator
	if all {
		iter = collection.Documents(ctx)
	} else if inactive {
		iter = collection.Where("inactive", "==", true).Documents(ctx)
	} else {
		iter = collection.Where("inactive", "!=", true).Documents(ctx)
	}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			err = common.ConvertGRPCError(err)
			return nil, err
		}

		d := Document{}
		if err := doc.DataTo(&d); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to read profile data", fid, "error", err)
			return nil, err
		}

		if d.SelectedCharacter == "" {
			d.SelectedCharacter = "2-xl"
		}

		docs = append(docs, d)
	}

	return docs, nil
}

// GetByID returns a profile given an ID
func GetByID(ctx context.Context, logCtx *slog.Logger, id string) (Document, error) {
	fid := slog.String("fid", "vox.profiles.GetByID")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	d := Document{}

	path := fmt.Sprintf("accounts/%s/profiles", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("profiles collection not found", fid)
		return d, common.ErrNotFound{}
	}

	doc, err := collection.Doc(id).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get profile document", fid, "error", err)
		return d, err
	}

	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read profile data", fid, "error", err)
		return d, err
	}

	if d.SelectedCharacter == "" {
		d.SelectedCharacter = "2-xl"
	}

	return d, nil
}

// GetByName returns a profile given a name
func GetByName(ctx context.Context, logCtx *slog.Logger, name string) ([]Document, error) {
	fid := slog.String("fid", "vox.profiles.GetByName")

	user := ctx.Value(common.AccountKey).(accounts.Document)

	docs := []Document{}

	path := fmt.Sprintf("accounts/%s/profiles", user.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("profiles collection not found", fid)
		return nil, common.ErrNotFound{}
	}

	iter := collection.Where("name", "==", name).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			err = common.ConvertGRPCError(err)
			return nil, err
		}

		d := Document{}
		if err := doc.DataTo(&d); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to read profile data", fid, "error", err)
			return nil, err
		}

		if d.SelectedCharacter == "" {
			d.SelectedCharacter = "2-xl"
		}

		docs = append(docs, d)
	}

	return docs, nil
}

// GetIDsByAccount returns all profiles IDs for a given account
func GetIDsByAccount(ctx context.Context, logCtx *slog.Logger, id string) ([]string, error) {
	fid := slog.String("fid", "vox.profiles.GetIDsByAccount")

	path := fmt.Sprintf("accounts/%s/profiles", id)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("profiles collection not found", fid)
		return nil, common.ErrNotFound{}
	}

	ids := []string{}

	iter := collection.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		ids = append(ids, doc.Ref.ID)
	}

	return ids, nil
}

// Patch makes a document inactive
func Patch(ctx context.Context, logCtx *slog.Logger, profileID string, document PatchDocument) (Document, error) {
	fid := slog.String("fid", "vox.profiles.Patch")

	user := ctx.Value(common.AccountKey).(accounts.Document)

	if profileID == "" {
		return Document{}, common.ErrBadRequest{}
	}

	path := fmt.Sprintf("accounts/%s/profiles", user.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("profiles collection not found", fid)
		return Document{}, common.ErrNotFound{}
	}

	update := []fs.Update{}

	if document.Name != nil {
		update = append(update, fs.Update{Path: "name", Value: *document.Name})

		res := moderate.Get(ctx, logCtx, *document.Name, "en-US")

		for _, v := range res.Categories {
			if v {
				return Document{}, common.ErrBadRequest{Src: "profile", Msg: "invalid profile name"}
			}
		}

		if res.Analysis.Toxic {
			return Document{}, common.ErrBadRequest{Src: "profile", Msg: "invalid profile name"}
		}
	}

	if document.AddQuestionFrequency != nil {
		update = append(update, fs.Update{Path: "add_question_frequency", Value: *document.AddQuestionFrequency})
	}

	if document.DontSay != nil {
		update = append(update, fs.Update{Path: "dont_say", Value: *document.DontSay})
	}

	if document.Inactive != nil {
		update = append(update, fs.Update{Path: "inactive", Value: *document.Inactive})
	}

	if document.Interests != nil {
		update = append(update, fs.Update{Path: "interests", Value: *document.Interests})
	}

	if document.Moderate != nil {
		update = append(update, fs.Update{Path: "moderate", Value: *document.Moderate})
	}

	if document.Notifications != nil {
		update = append(update, fs.Update{Path: "notifications", Value: *document.Notifications})
	}

	if document.ReplaceWords != nil {
		for k, v := range *document.ReplaceWords {
			if v != nil {
				update = append(update, fs.Update{Path: "replace_words." + k, Value: v})
			} else {
				update = append(update, fs.Update{Path: "replace_words." + k, Value: fs.Delete})
			}
		}
	}

	if document.ResponseAge != nil {
		update = append(update, fs.Update{Path: "response_age", Value: *document.ResponseAge})
	}

	if document.SelectedCharacter != nil {
		update = append(update, fs.Update{Path: "selected_character", Value: *document.SelectedCharacter})
	}

	if document.TopicsDiscourage != nil {
		update = append(update, fs.Update{Path: "topics_discourage", Value: *document.TopicsDiscourage})
	}

	if document.TopicsEncourage != nil {
		update = append(update, fs.Update{Path: "topics_encourage", Value: *document.TopicsEncourage})
	}

	if len(update) > 0 {
		update = append(update, fs.Update{Path: "modified_date", Value: time.Now()})

		if _, err := collection.Doc(profileID).Update(ctx, update); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to update profile document", fid, "error", err)
			return Document{}, err
		}
	}

	p, err := GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Warn("unable to get profile document", fid, "error", err)
		return Document{}, err
	}

	return p, nil
}

// PatchCharacter patches a profile's character preferences
func PatchCharacter(ctx context.Context, logCtx *slog.Logger, profileID, characterName string, document PatchCharacterPreferences) (Document, error) {
	fid := slog.String("fid", "vox.profiles.Patch")

	user := ctx.Value(common.AccountKey).(accounts.Document)

	if profileID == "" {
		logCtx.Error("profile_id required", fid)
		return Document{}, common.ErrBadRequest{}
	}

	path := fmt.Sprintf("accounts/%s/profiles", user.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("profiles collection not found", fid)
		return Document{}, common.ErrNotFound{}
	}

	update := []fs.Update{}

	if document.ImageStyle != nil {
		update = append(update, fs.Update{Path: "characters." + characterName + ".image_style", Value: *document.ImageStyle})
	}

	if document.Language != nil {
		update = append(update, fs.Update{Path: "characters." + characterName + ".language", Value: *document.Language})
	}

	if document.Mode != nil {
		update = append(update, fs.Update{Path: "characters." + characterName + ".mode", Value: *document.Mode})
	}

	if document.Voice != nil {
		update = append(update, fs.Update{Path: "characters." + characterName + ".voice", Value: *document.Voice})
	}

	if len(update) > 0 {
		if _, err := collection.Doc(profileID).Update(ctx, update); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to update profile document", fid, "error", err)
			return Document{}, err
		}
	}

	p, err := GetByID(ctx, logCtx, profileID)
	if err != nil {
		logCtx.Warn("unable to get profile document", fid, "error", err)
		return Document{}, err
	}

	return p, nil
}
