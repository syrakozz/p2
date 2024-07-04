package characters

import (
	"context"
	"log/slog"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/firestore"
)

// CharacterInfo contains information about a character.
type CharacterInfo struct {
	Modes  []string          `firestore:"modes,omitempty" json:"modes,omitempty"`
	Voices map[string]string `firestore:"voices,omitempty" json:"voices,omitempty"`
}

// GetCharacter retrieves a character from firestore.
func GetCharacter(ctx context.Context, logCtx *slog.Logger, characterVersion, language string) (Character, error) {
	fid := slog.String("fid", "vox.characters.GetCharacter")

	if language == "" {
		language = "en-US"
	}

	characterLanguage := characterVersion + "_" + language

	if c, ok := characters[characterLanguage]; ok && !config.VARS.DisableCaches {
		return c, nil
	}

	collection := firestore.Client.Collection("characters")
	if collection == nil {
		logCtx.Error("characters collection not found", fid)
		return Character{}, common.ErrNotFound{}
	}

	doc, err := collection.Doc(characterLanguage).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get character document", fid, "error", err)
		return Character{}, err
	}

	c := Character{}

	if err := doc.DataTo(&c); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read character data", fid, "error", err)
		return Character{}, err
	}

	// memoize
	characters[characterLanguage] = c

	return c, nil
}

// SetCharacter updates a character in firestore.
func SetCharacter(ctx context.Context, logCtx *slog.Logger, characterVersion, language string, character Character) error {
	fid := slog.String("fid", "vox.characters.SetCharacter")

	if language == "" {
		language = "en-US"
	}

	collection := firestore.Client.Collection("characters")
	if collection == nil {
		logCtx.Error("characters collection not found", fid)
		return common.ErrNotFound{}
	}

	if _, err := collection.Doc(characterVersion+"_"+language).Set(ctx, character); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to set character document", fid, "error", err)
		return err
	}

	return nil
}

// SetCharacters updates a list of characters in firestore.
func SetCharacters(ctx context.Context, logCtx *slog.Logger, characters map[string]Character) error {
	fid := slog.String("fid", "vox.characters.SetCharacters")

	collection := firestore.Client.Collection("characters")
	if collection == nil {
		logCtx.Error("characters collection not found", fid)
		return common.ErrNotFound{}
	}

	for name, character := range characters {
		if _, err := collection.Doc(name).Set(ctx, character); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to set character document", fid, "error", err)
			return err
		}
	}

	return nil
}
