package openai

import (
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/openai"
	"disruptive/lib/users"
	e "disruptive/rest/errors"
)

func postEmbeddings(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.openai.postEmbeddings")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"ai.user"})
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	req := openai.EmbeddingsRequest{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read embeddings data")
	}

	if req.Model == "" {
		req.Model = "text-embedding-ada-002"
	}

	embeddings, err := openai.PostEmbeddings(ctx, logCtx, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process query")
	}

	logCtx.Info("post embeddings", fid)
	return c.JSON(http.StatusOK, embeddings)
}

func postEmbeddingsSimilarity(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.openai.postEmbeddingsSimilarity")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"ai.user"})
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	req := openai.EmbeddingsRequest{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read embeddings data")
	}

	if req.Model == "" {
		req.Model = "text-embedding-ada-002"
	}

	if len(req.Input) != 2 {
		return e.ErrBad(logCtx, fid, "input must contain exactly 2 values")
	}

	similarity, err := openai.PostEmbeddingsSimilarity(ctx, logCtx, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process query")
	}

	logCtx.Info("post embeddings", fid)
	return c.JSON(http.StatusOK, map[string]float64{"similarity": similarity})
}
