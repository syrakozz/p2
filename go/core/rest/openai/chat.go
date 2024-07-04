package openai

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/openai"
	"disruptive/lib/users"
	e "disruptive/rest/errors"
)

func postChat(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.openai.postChat")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"ai.user"})
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	req := openai.ChatRequest{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read chat data")
	}

	logCtx = logCtx.With("username", claims.Username)

	stream := c.QueryParam("stream") == "1"
	charsRequest := 0
	charsResponse := 0

	for _, m := range req.Messages {
		charsRequest += len(m.Content)
	}

	// Return steam response
	if stream {
		r, err := openai.PostChatStream(ctx, logCtx, req)
		if err != nil {
			return e.Err(logCtx, err, fid, "unable to post chat")
		}

		c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
		c.Response().WriteHeader(http.StatusOK)

		buf := make([]byte, 16)

		for {
			n, err := r.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				return e.Err(logCtx, err, fid, "unable to read buffer")
			}

			c.Response().Write(buf[:n])
			c.Response().Flush()

			charsResponse += n
		}

		logCtx.Info("tokens", "request", charsRequest/4, "response", charsResponse/4)
		return nil
	}

	// Return single response
	res, err := openai.PostChat(ctx, logCtx, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to post chat")
	}

	logCtx.Info("chat", fid)
	return c.JSON(http.StatusOK, res)

}
