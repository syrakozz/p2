package openai

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/http"

	"disruptive/lib/common"
)

// EmbeddingsRequest is a the request API structure.
type EmbeddingsRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingsResponse is the respone API structure.
type EmbeddingsResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	}
}

// PostEmbeddings sends embededdings to OpenAI.
func PostEmbeddings(ctx context.Context, logCtx *slog.Logger, req EmbeddingsRequest) (EmbeddingsResponse, error) {
	fid := slog.String("fid", "openai.PostEmbeddings")

	if err := textAda002Limiter.Wait(ctx); err != nil {
		return EmbeddingsResponse{}, err
	}

	res, err := Resty.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&EmbeddingsResponse{}).
		Post(embeddingsEndpoint)

	if err != nil {
		logCtx.Error("openai embeddings endpoint failed", fid, "error", err)
		return EmbeddingsResponse{}, err
	}

	if res.StatusCode() == http.StatusUnauthorized {
		logCtx.Error("openai unauthorized", fid, "status", res.Status())
		return EmbeddingsResponse{}, common.ErrUnauthorized
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("openai embeddings endpoint failed", fid, "status", res.Status(), "message", errorMessage(res.Body()))
		return EmbeddingsResponse{}, errors.New(res.Status())
	}

	return *res.Result().(*EmbeddingsResponse), nil
}

// CosineSimilarity calculates the difference between 2 vectors.
func CosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0.0, common.ErrBadRequest{Msg: "both vectors must be the same length"}
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += math.Pow(a[i], 2)
		normB += math.Pow(b[i], 2)
	}

	if normA == 0 || normB == 0 {
		return 0.0, nil
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)), nil
}

// PostEmbeddingsSimilarity sends embededdings to OpenAI.
func PostEmbeddingsSimilarity(ctx context.Context, logCtx *slog.Logger, req EmbeddingsRequest) (float64, error) {
	fid := slog.String("fid", "openai.PostEmbeddingsSimilarity")

	if len(req.Input) != 2 {
		logCtx.Error("input must be length of 2")
		return 0.0, common.ErrBadRequest{Msg: "input must be length of 2"}
	}

	res, err := PostEmbeddings(ctx, logCtx, req)
	if err != nil {
		logCtx.Error("unable to post embeddings", fid, "error", err)
		return 0.0, err
	}

	if len(res.Data) != 2 {
		logCtx.Error("response data must be length of 2", fid)
		return 0.0, common.ErrConsistency
	}

	similarity, err := CosineSimilarity(res.Data[0].Embedding, res.Data[1].Embedding)
	if err != nil {
		logCtx.Error("unable to calculate cosine similarity", fid, "error", err)
		return 0.0, err
	}

	return similarity, nil
}
