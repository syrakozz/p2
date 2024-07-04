// Package demo ...
package demo

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/pkg/vox/demo"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// PatchUser update a demo user document.s
func PatchUser(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.demo.GetPatchUser")

	user := c.Param("user")

	req := demo.Document{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if err := demo.PatchUser(ctx, logCtx, user, req); err != nil {
		return e.Err(logCtx, err, fid, "unable to patch demo user document")
	}

	return c.NoContent(http.StatusOK)
}
