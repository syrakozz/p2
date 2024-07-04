package configs

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"

	fs "cloud.google.com/go/firestore"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/firestore"
)

// Localize contains the localization config.
type Localize struct {
	Character    map[string]string `firestore:"character" json:"character"`
	Email        map[string]string `firestore:"email" json:"email"`
	Moderation   map[string]string `firestore:"moderation" json:"moderation"`
	Predefined   map[string]string `firestore:"predefined" json:"predefined"`
	TextAnalysis map[string]string `firestore:"text_analysis" json:"text_analysis"`
}

// Localizations is a sync.Map for Localize
var Localizations sync.Map

// GetLocalization returns the localize character configs.
func GetLocalization(ctx context.Context, logCtx *slog.Logger, version, language string) (Localize, error) {
	fid := slog.String("fid", "console.configs.GetLocalization")

	if language == "" {
		language = "en"
	}

	if a, ok := Localizations.Load(language); ok && !config.VARS.DisableCaches {
		return a.(Localize), nil
	}

	collection := firestore.Client.Collection("configs")
	if collection == nil {
		logCtx.Warn("configs collection not found", fid)
		return Localize{}, common.ErrNotFound{}
	}

	var (
		doc *fs.DocumentSnapshot
		err error
	)

	doc, err = collection.Doc(strings.Join([]string{"localize", version, language}, "_")).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		if !errors.Is(err, common.ErrNotFound{}) {
			logCtx.Error("unable to get character localize config", fid, "error", err)
			return Localize{}, err
		}
	}

	if !doc.Exists() {
		doc, err = collection.Doc(strings.Join([]string{"localize", version, "en-US"}, "_")).Get(ctx)
		if err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to get character localize config", fid, "error", err)
			return Localize{}, err
		}
	}

	l := Localize{}

	if err := doc.DataTo(&l); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read configs localize data", fid, "error", err)
		return Localize{}, err
	}

	Localizations.Store(language, l)

	return l, nil
}

// SetLocalization imports the given list of localization config files to firestore.
func SetLocalization(ctx context.Context, logCtx *slog.Logger, localizations map[string]Localize) error {
	fid := slog.String("fid", "vox.configs.SetLocalization")

	collection := firestore.Client.Collection("configs")
	if collection == nil {
		logCtx.Error("configs collection not found", fid)
		return common.ErrNotFound{}
	}

	for name, localize := range localizations {
		if _, err := collection.Doc(name).Set(ctx, localize); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to set config localize document", fid, "name", name, "error", err)
			return err
		}
	}

	return nil
}
