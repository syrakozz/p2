package configs

import (
	"context"
	"log/slog"

	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/firestore"
)

// Document is the config data.
type Document map[string]any

// Get returns the services config data.
func Get(ctx context.Context, logCtx *slog.Logger, document string) (Document, error) {
	logCtx = logCtx.With("fid", "configs.Get")

	doc, err := firestore.Client.Collection("configs").Doc(document).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		if config.VARS.Env == "local" {
			logCtx.Warn("unable to get config", "error", err)
		}
		return nil, err
	}

	d := Document{}
	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read config data", "error", err)
		return nil, err
	}

	return d, nil
}

// Put replaces a config file.
func Put(ctx context.Context, logCtx *slog.Logger, name string, document Document) error {
	logCtx = logCtx.With("fid", "configs.Put")

	if _, err := firestore.Client.Collection("configs").Doc(name).Set(ctx, document); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to set config", "error", err)
		return err
	}

	return nil
}
