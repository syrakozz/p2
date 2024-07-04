package sage

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/sage/models/chatpad"
	u "disruptive/lib/users"
	e "disruptive/rest/errors"
)

func postModelChatpad(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.postModelChatpad")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := u.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), sageUser)
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := claims.Username
	logCtx = logCtx.With("username", username)

	logCtx, project, err := getUUIDParam(c, logCtx, "_project")
	if err != nil {
		return err
	}

	logCtx, session, err := getUUIDParam(c, logCtx, "_session")
	if err != nil {
		return err
	}

	logCtx, model, err := getUUIDParam(c, logCtx, "_model")
	if err != nil {
		return err
	}

	stream := c.QueryParam("stream") == "1"

	req := struct {
		Message string `json:"message"`
	}{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if stream {
		r, tokensPrompt, entry, err := chatpad.PostStream(ctx, logCtx, username, project, session, model, req.Message)
		if err != nil {
			return e.Err(logCtx, err, fid, "unable to create model")
		}

		c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
		c.Response().Header().Set("X-Sage-Entry", fmt.Sprintf("%04d", entry+1))
		c.Response().WriteHeader(http.StatusOK)

		// support 32k
		bufAll := bytes.NewBuffer(make([]byte, 0, 32*1024))
		bufToken := make([]byte, 16)
		tokensResponse := 0

		for {
			// each read is 1 token
			n, err := r.Read(bufToken)
			if err != nil {
				if err == io.EOF {
					break
				}
				return e.Err(logCtx, err, fid, "unable to read buffer")
			}

			bufAll.Write(bufToken[:n])
			c.Response().Write(bufToken[:n])
			c.Response().Flush()

			tokensResponse++
		}

		if err := chatpad.AddGPTAssistantEntry(ctx, logCtx, username, project, session, model, bufAll.String(), tokensPrompt, tokensResponse); err != nil {
			return e.Err(logCtx, err, fid, "unable to add assistant entry")
		}

		logCtx.Info("tokens", fid, "prompt", tokensPrompt, "response", tokensResponse)
		return nil
	}

	// Return single response
	res, err := chatpad.Post(ctx, logCtx, username, project, session, model, req.Message)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to post chat")
	}

	logCtx.Info("chatpad", fid)
	return c.JSON(http.StatusOK, res)
}

func postModelChatpadStreamFinalize(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.sage.postModelChatpadStreamFinalize")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	claims := u.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), sageUser)
	if !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	username := claims.Username
	logCtx = logCtx.With("username", username)

	logCtx, project, err := getUUIDParam(c, logCtx, "_project")
	if err != nil {
		return err
	}

	logCtx, session, err := getUUIDParam(c, logCtx, "_session")
	if err != nil {
		return err
	}

	logCtx, model, err := getUUIDParam(c, logCtx, "_model")
	if err != nil {
		return err
	}

	res, err := chatpad.CreateResponse(ctx, logCtx, username, project, session, model)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to create reponse")
	}

	logCtx.Info("stream finalized", fid)
	return c.JSON(http.StatusOK, res)
}
