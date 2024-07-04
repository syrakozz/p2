package configs

import (
	"context"
	"errors"
	"log/slog"

	"disruptive/lib/common"
	"disruptive/lib/firestore"
)

// StableDiffusion contains the Stable Diffusion configuration.
type StableDiffusion struct {
	NegativePrompts []string `firestore:"negative_prompts" json:"negative_prompts"`
	Styles          []string `firestore:"styles" json:"styles"`
}

var stableDiffusion *StableDiffusion

// GetStableDiffusion returns the localize character configs.
func GetStableDiffusion(ctx context.Context, logCtx *slog.Logger) (*StableDiffusion, error) {
	fid := slog.String("fid", "console.configs.GetStableDiffusion")

	if stableDiffusion != nil {
		return stableDiffusion, nil
	}

	collection := firestore.Client.Collection("configs")
	if collection == nil {
		logCtx.Warn("configs collection not found", fid)
		return nil, common.ErrNotFound{}
	}

	doc, err := collection.Doc("stablediffusion").Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		if !errors.Is(err, common.ErrNotFound{}) {
			logCtx.Error("unable to get stablediffusion config", fid, "error", err)
			return nil, err
		}
	}

	stableDiffusion = &StableDiffusion{}

	if err := doc.DataTo(stableDiffusion); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read configs stablediffusion data", fid, "error", err)
		return nil, err
	}

	return stableDiffusion, nil
}
