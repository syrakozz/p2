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

func postQuery(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.pinecone.postQuery")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	if claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"ai.user"}); !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	req := pinecone.QueryRequest{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if req.Namespace == "" {
		return e.ErrBad(logCtx, fid, "missing namespace")
	}

	if req.ID == "" && len(req.Vector) == 0 {
		return e.ErrBad(logCtx, fid, "missing id or vector")
	}

	if req.TopK < 1 {
		req.TopK = 1
	}

	req.IncludeMetadata = true

	res, err := pinecone.Query(ctx, logCtx, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process query")
	}

	logCtx.Info("processed query", fid)
	return c.JSON(http.StatusOK, res)
}
