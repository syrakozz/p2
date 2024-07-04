package characters

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"disruptive/pkg/vox/characters"
)

// Import sets a character in firestore from a file.
func Import(ctx context.Context, characterVersion, language string) error {
	logCtx := slog.With("fid", "console.vox.characters.Import")

	parts := strings.Split(characterVersion, "_")
	characterName := parts[0]
	version := parts[1]

	cf, err := os.ReadFile(fmt.Sprintf("data/characters/%s/%s/%s_%s.json", characterName, version, characterVersion, language))
	if err != nil {
		logCtx.Error("unable to read character json file", "error", err)
		return err
	}

	c := characters.Character{}
	if err = json.Unmarshal(cf, &c); err != nil {
		logCtx.Error("unable to unmarshal json", "error", err)
		return err
	}

	if err := characters.SetCharacter(ctx, logCtx, characterVersion, language, c); err != nil {
		logCtx.Error("unable to set character", "error", err, "character", c)
		return err
	}

	return nil
}

// ImportAll sets all characters with the same characterName and version into firestore.
func ImportAll(ctx context.Context, characterVersion string) error {
	logCtx := slog.With("fid", "console.vox.characters.ImportAll")

	parts := strings.Split(characterVersion, "_")
	characterName := parts[0]
	version := parts[1]

	files, err := os.ReadDir(fmt.Sprintf("data/characters/%s/%s", characterName, version))
	if err != nil {
		logCtx.Error("unable to read data/characters directory", "error", err)
		return err
	}

	cList := make(map[string]characters.Character, len(files))
	for _, f := range files {
		cf, err := os.ReadFile(fmt.Sprintf("data/characters/%s/%s/%s", characterName, version, f.Name()))
		if err != nil {
			logCtx.Error("unable to read character json file", "error", err)
			return err
		}

		c := characters.Character{}
		if err = json.Unmarshal(cf, &c); err != nil {
			logCtx.Error("unable to unmarshal json", "error", err)
			return err
		}

		cList[strings.Split(f.Name(), ".")[0]] = c
	}

	if len(cList) == 0 {
		logCtx.Info("no files found")
		return nil
	}

	if err := characters.SetCharacters(ctx, logCtx, cList); err != nil {
		logCtx.Error("unable to set characters", "error", err, "character", characterName, "version", version)
		return err
	}

	return nil
}
