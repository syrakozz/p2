package configs

import (
	"context"
	"log/slog"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/elevenlabs"
	"disruptive/lib/firestore"
)

// SetElevenlabsVoices returns elevenlabs voices.
func SetElevenlabsVoices(ctx context.Context, logCtx *slog.Logger) error {
	logCtx = logCtx.With("fid", "configs.GetElevenlabsVoices")

	doc, err := firestore.Client.Collection("configs").Doc("elevenlabs_voices_v2").Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		if config.VARS.Env == "local" {
			logCtx.Warn("unable to get config", "error", err)
		}
		return err
	}

	if err := doc.DataTo(&elevenlabs.Voices); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read config data", "error", err)
		return err
	}

	return nil
}
