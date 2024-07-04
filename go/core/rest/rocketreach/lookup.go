package rocketreach

import (
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/rocketreach"
	"disruptive/lib/users"
	e "disruptive/rest/errors"
)

func getLookup(c echo.Context) error {
	ctx := c.Request().Context()
	fid := slog.String("fid", "rest.rocketreach.getLookup")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	if claims := users.VerifyClaims(ctx, logCtx, c.Get("user").(*jwt.Token), []string{"disruptive.service", "rocketreach.user"}); !claims.IsVerified {
		return e.Err(logCtx, common.ErrUnauthorized, fid, "unauthorized")
	}

	id := c.QueryParam("id")
	name := c.QueryParam("name")
	currentEmployer := c.QueryParam("current_employer")
	title := c.QueryParam("title")
	linkedInURL := c.QueryParam("linkedin_url")
	email := c.QueryParam("email")

	if (name == "" && currentEmployer != "") || (name != "" && currentEmployer == "") {
		return e.ErrBad(logCtx, fid, "name and current_employer must both be emtpy or both have values")
	}

	req := rocketreach.LookupRequest{
		ID:              id,
		Name:            name,
		CurrentEmployer: currentEmployer,
		Title:           title,
		LinkedInURL:     linkedInURL,
		Email:           email,
	}

	lookup, err := rocketreach.GetLookup(ctx, logCtx, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get lookup")
	}

	logCtx.Info("get lookup", fid)
	return c.JSON(http.StatusOK, lookup)
}
