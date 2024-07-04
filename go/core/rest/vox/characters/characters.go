// Package characters handles character APIs.
package characters

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/lib/configs"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetCharacters returns modes and voices for all characters.
func GetCharacters(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.characters.GetCharacters")

	characters, err := configs.Get(ctx, logCtx, "characters")
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get character infos")
	}

	return c.JSON(http.StatusOK, characters)
}
