package mailfinder

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/lib/mailfinder"
	e "disruptive/rest/errors"
)

func postSearchEmployees(c echo.Context) error {
	fid := slog.String("fid", "rest.mailfinder.postSearchEmployees")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	req := struct {
		Domain  string `json:"domain"`
		Company string `json:"company"`
		Title   string `json:"title"`
		Country string `json:"country"`
	}{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read search data")
	}

	logCtx = logCtx.With("domain", req.Domain, "company", req.Company, "title", req.Title, "country", req.Country)

	if req.Domain == "" && req.Company == "" {
		return e.ErrBad(logCtx, fid, "missing domain or company")
	}

	if req.Title == "" {
		return e.ErrBad(logCtx, fid, "missing title")
	}

	results, err := mailfinder.PostSearchEmployees(c.Request().Context(), logCtx, req.Domain, req.Company, req.Title, req.Country)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to retrieve search employees result")
	}

	logCtx.Info("retrieved search employees result", fid)
	return c.JSON(http.StatusOK, results)
}
