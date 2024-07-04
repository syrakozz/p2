package configs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"disruptive/lib/configs"
)

// Import sets a config in firestore from a file.
func Import(ctx context.Context, configName string) error {
	logCtx := slog.With("fid", "console.console.configs.Import")

	filePath := "data/configs/" + configName + ".json"
	if strings.Contains(configName, "localize") {
		version := strings.Split(configName, "_")[1]
		filePath = fmt.Sprintf("data/configs/localize/%s/%s.json", version, configName)
	}

	cf, err := os.ReadFile(filePath)
	if err != nil {
		logCtx.Error("unable to read config json file", "error", err)
		return err
	}

	c := configs.Document{}
	if err = json.Unmarshal(cf, &c); err != nil {
		logCtx.Error("unable to unmarshal json", "error", err)
		return err
	}

	if err := configs.Put(ctx, logCtx, configName, c); err != nil {
		logCtx.Error("unable to set config", "error", err, "config", c)
		return err
	}

	return nil
}

// ImportAllLocalizations sets all localization files with the same name and version into firestore.
func ImportAllLocalizations(ctx context.Context, configName string) error {
	logCtx := slog.With("fid", "console.vox.configs.ImportAllLocalizations")

	parts := strings.Split(configName, "_")
	version := parts[1]

	files, err := os.ReadDir(fmt.Sprintf("data/configs/localize/%s", version))
	if err != nil {
		logCtx.Error("unable to read data/characters directory", "error", err)
		return err
	}

	cList := make(map[string]configs.Localize, len(files))
	for _, f := range files {
		cf, err := os.ReadFile(fmt.Sprintf("data/configs/localize/%s/%s", version, f.Name()))
		if err != nil {
			logCtx.Error("unable to read character json file", "error", err)
			return err
		}

		c := configs.Localize{}
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

	if err := configs.SetLocalization(ctx, logCtx, cList); err != nil {
		logCtx.Error("unable to set localization files", "error", err, "version", version)
		return err
	}

	return nil
}
