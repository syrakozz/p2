package pinecone

import (
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/pinecone"
	"disruptive/lib/users"
	e "disruptive/rest/errors"
)

func postDocuments(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.pinecone.postDocuments")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	if claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"ai.user"}); !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	req := pinecone.UpsertRequest{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if req.Namespace == "" {
		return e.ErrBad(logCtx, fid, "invalid namespace")
	}

	if len(req.Vectors) == 0 {
		return e.ErrBad(logCtx, fid, "invalid vectors")
	}

	ids, err := pinecone.Upsert(ctx, logCtx, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process upsert")
	}

	logCtx.Info("processed upsert", fid)
	return c.JSON(http.StatusOK, map[string][]string{"ids": ids})
}

func getDocuments(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.pinecone.getDocuments")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	if claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"ai.user"}); !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	q := c.QueryParams()
	ns := q.Get("namespace")
	ids := q.Get("ids")

	if ns == "" {
		return e.ErrBad(logCtx, fid, "invalid namespace")
	}

	if ids == "" {
		return e.ErrBad(logCtx, fid, "invalid ids")
	}

	req := pinecone.FetchRequest{
		Namespace: ns,
		IDs:       q["ids"],
	}

	vectors, err := pinecone.Fetch(ctx, logCtx, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get document")
	}

	logCtx.Info("get document", fid)
	return c.JSON(http.StatusOK, vectors)
}

func patchDocuments(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.pinecone.patchDocuments")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	if claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"ai.user"}); !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	req := pinecone.UpdateRequest{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if req.Namespace == "" {
		return e.ErrBad(logCtx, fid, "invalid namespace")
	}

	if err := pinecone.Update(ctx, logCtx, req); err != nil {
		return e.Err(logCtx, err, fid, "unable to update document")
	}

	logCtx.Info("updated document", fid)
	return c.NoContent(http.StatusOK)
}

func deleteDocuments(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.pinecone.deleteDocuments")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	if claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"ai.user"}); !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	req := pinecone.DeleteRequest{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if req.Namespace == "" {
		return e.ErrBad(logCtx, fid, "invalid namespace")
	}

	if !req.DeleteAll && len(req.IDs) == 0 && len(req.Filter) == 0 {
		return e.ErrBad(logCtx, fid, "nothing to delete")
	}

	if err := pinecone.Delete(ctx, logCtx, req); err != nil {
		return e.Err(logCtx, err, fid, "unable to delete documents")
	}

	logCtx.Info("deleted docmuents", fid)
	return c.NoContent(http.StatusOK)
}
