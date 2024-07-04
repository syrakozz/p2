package llm

import (
	"context"
	"log/slog"

	"disruptive/lib/common"
	"disruptive/lib/openai"
)

// ListModels lists OpenAI finetune models and includes basic information.
func ListModels(ctx context.Context) error {
	logCtx := slog.With()

	res, err := openai.GetFinetunes(ctx, logCtx)
	if err != nil {
		return err
	}

	common.P(res)
	return nil
}

// ListModel displays detail information for an OpenAI finetune model.
func ListModel(ctx context.Context, modelID string) error {
	logCtx := slog.With()

	res, err := openai.GetFinetune(ctx, logCtx, modelID)
	if err != nil {
		return err
	}

	common.P(res)
	return nil
}

// CreateModel creates a new finetune OpenAI model.
func CreateModel(ctx context.Context, fileID, model string, nEpochs int, suffix string) error {
	logCtx := slog.With()

	req := openai.FinetuneRequest{
		TrainingFile: fileID,
		Model:        model,
		NEpochs:      nEpochs,
		Suffix:       suffix,
	}

	res, err := openai.PostFinetune(ctx, logCtx, req)
	if err != nil {
		return err
	}

	common.P(res)
	return nil
}

// DeleteModel delete an OpenAI finetune model.
func DeleteModel(ctx context.Context, modelID string) error {
	logCtx := slog.With()
	return openai.DeleteModel(ctx, logCtx, modelID)
}
