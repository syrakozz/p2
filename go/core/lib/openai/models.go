package openai

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"disruptive/lib/common"
)

// ModelResponse is the response API structure.
type ModelResponse struct {
	ID      string `json:"id"`
	OwnedBy string `json:"owned_by"`
}

type modelResponse struct {
	Data []ModelResponse `json:"data"`
}

type deleteModelResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// GetModels returns all available OpenAI models.
func GetModels(ctx context.Context, logCtx *slog.Logger) ([]ModelResponse, error) {
	fid := slog.String("fid", "openai.GetModels")

	res, err := Resty.R().
		SetContext(ctx).
		SetResult(&modelResponse{}).
		Get(modelsEndpoint)

	if err != nil {
		logCtx.Error("openai models endpoint failed", fid, "error", err)
		return nil, err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("openai unauthorized", fid, "status", res.Status())
		return nil, common.ErrUnauthorized
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("openai model endpoint failed", fid, "status", res.Status())
		return nil, errors.New(res.Status())
	}

	return res.Result().(*modelResponse).Data, nil
}

// DeleteModel deletes an OpenAI model.
func DeleteModel(ctx context.Context, logCtx *slog.Logger, modelID string) error {
	fid := slog.String("fid", "openai.DeleteModel")

	res, err := Resty.R().
		SetContext(ctx).
		SetPathParam("model_id", modelID).
		SetResult(&deleteModelResponse{}).
		Delete(modelEndpoint)

	if err != nil {
		logCtx.Error("openai model endpoint failed", fid, "error", err)
		return err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("openai unauthorized", fid, "status", res.Status())
		return common.ErrUnauthorized
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("openai model endpoint failed", fid, "status", res.Status())
		return errors.New(res.Status())
	}

	if res.Result().(*deleteModelResponse).ID != modelID || !res.Result().(*deleteModelResponse).Deleted {
		logCtx.Error("unable to delete model", fid, "model_id", modelID)
		return errors.New("unable to delete model")
	}

	return nil
}
