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

func postStats(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.pinecone.postStats")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	if claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"disruptive.service", "ai.user"}); !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	req := pinecone.Metadata{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	stats, err := pinecone.Stats(ctx, logCtx, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to process stats")
	}

	logCtx.Info("processed stats", fid)
	return c.JSON(http.StatusOK, stats)
}
