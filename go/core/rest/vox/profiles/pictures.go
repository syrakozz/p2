package profiles

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"disruptive/pkg/vox/profiles"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetProfilePicture gets a profile's picture.
func GetProfilePicture(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.GetProfilePicture")

	profileID := c.Param("profile_id")

	r, cType, err := profiles.GetProfilePicture(ctx, logCtx, profileID)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get profile picture")
	}
	defer r.Close()

	return c.Stream(http.StatusOK, cType, r)
}

// PutProfilePicture updates a profile's picture.
func PutProfilePicture(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.PutProfilePicture")

	profileID := c.Param("profile_id")

	file, err := c.FormFile("file")
	if err != nil {
		return e.ErrBad(logCtx, fid, "invalid file")
	}

	filename := strings.ToLower(file.Filename)

	f, err := file.Open()
	if err != nil {
		return e.ErrBad(logCtx, fid, "unable to open file")
	}
	defer f.Close()

	if err := profiles.PutProfilePicture(ctx, logCtx, profileID, filename, f); err != nil {
		return e.Err(logCtx, err, fid, "unable to update profile picture")
	}

	return c.NoContent(http.StatusOK)
}
