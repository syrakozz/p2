package configs

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"disruptive/lib/configs"
)

// Export saves a config in firestore to a file.
func Export(ctx context.Context, configName string) error {
	logCtx := slog.With("fid", "console.console.configs.Export")

	var (
		data any
		err  error
	)

	data, err = configs.Get(ctx, logCtx, configName)
	if err != nil {
		logCtx.Error("unable to export config", "error", err, "config", configName)
		return err
	}

	jsonC, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		logCtx.Error("unable to marshall config to JSON", "error", err)
		return err
	}
	jsonC = append(jsonC, '\n')

	if err = os.WriteFile("data/configs/"+configName+".json", jsonC, 0644); err != nil {
		logCtx.Error("unable to create and write to config file", "error", err)
		return err
	}

	return nil
}
