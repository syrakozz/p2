package profiles

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"disruptive/pkg/vox/characters"
	"disruptive/pkg/vox/profiles"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetAll returns all profiles for an account.
func GetAll(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.GetAll")

	name := strings.ToLower(c.QueryParam("name"))
	all := c.QueryParam("all") == "true"
	inactive := c.QueryParam("inactive") == "true"

	var (
		err error
	)

	if all && inactive {
		return e.ErrBad(logCtx, fid, "all and inactive cannot both be true")
	}

	var p []profiles.Document

	switch {
	case name != "":
		p, err = profiles.GetByName(ctx, logCtx, name)
	default:
		p, err = profiles.Get(ctx, logCtx, all, inactive)
	}

	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get profiles")
	}

	return c.JSON(http.StatusOK, p)
}

// Get returns a profile.
func Get(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.Get")

	id := c.Param("profile_id")

	p, err := profiles.GetByID(ctx, logCtx, id)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get profile")
	}

	return c.JSON(http.StatusOK, p)
}

// Post creates a new profile.
func Post(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.Post")

	d := profiles.Document{}

	if err := c.Bind(&d); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if d.Name == "" {
		return e.ErrBad(logCtx, fid, "invalid name")
	}

	doc, err := profiles.Create(ctx, logCtx, d)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to create profile")
	}

	return c.JSON(http.StatusOK, doc)
}

// Patch updates a profile.
func Patch(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.Patch")

	profileID := c.Param("profile_id")

	d := profiles.PatchDocument{}

	if err := c.Bind(&d); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	p, err := profiles.Patch(ctx, logCtx, profileID, d)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to patch profile")
	}

	return c.JSON(http.StatusOK, p)
}

// Delete sets a profile to inactive or full deletes a profile.
func Delete(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.Delete")

	profileID := c.Param("profile_id")

	b := true
	_, err := profiles.Patch(ctx, logCtx, profileID, profiles.PatchDocument{Inactive: &b})
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to set profile to inactive")
	}

	return c.NoContent(http.StatusNoContent)
}

// PatchCharacter updates a profile's character preferences.
func PatchCharacter(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.PatchCharacter")

	profileID := c.Param("profile_id")
	characterVersion := c.Param("character_version")

	d := profiles.PatchCharacterPreferences{}

	if err := c.Bind(&d); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	// Remove when UI passes in correct character_version
	if !strings.HasSuffix(characterVersion, "_v1") {
		characterVersion += "_v1"
	}

	characterName := strings.Split(characterVersion, "_")[0]

	p, err := profiles.PatchCharacter(ctx, logCtx, profileID, characterName, d)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to patch profile character")
	}

	if d.Mode != nil && *d.Mode != "" {
		profile, err := profiles.GetByID(ctx, logCtx, profileID)
		if err != nil {
			return e.Err(logCtx, err, fid, "invalid profile")
		}

		if _, err := characters.EndSequence(ctx, logCtx, &profile, characterVersion); err != nil {
			return e.Err(logCtx, err, fid, "unable to process end sequence")
		}
	}

	return c.JSON(http.StatusOK, p)
}
