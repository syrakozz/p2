package characters

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"disruptive/pkg/vox/characters"
)

// Export saves a character in firestore to a file.
func Export(ctx context.Context, characterVersion, language string) error {
	logCtx := slog.With("fid", "console.vox.characters.Export")

	var (
		data any
		err  error
	)

	data, err = characters.GetCharacter(ctx, logCtx, characterVersion, language)
	if err != nil {
		logCtx.Error("unable to export character", "error", err, "character", characterVersion)
		return err
	}

	jsonC, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		logCtx.Error("unable to marshall character to JSON", "error", err)
		return err
	}
	jsonC = append(jsonC, '\n')

	if err = os.WriteFile("data/characters/"+characterVersion+"_"+language+".json", jsonC, 0644); err != nil {
		logCtx.Error("unable to create and write to character file", "error", err)
		return err
	}

	return nil
}
