package profiles

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"disruptive/pkg/vox/characters"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetArchiveByID returns a profile archive by ID.
func GetArchiveByID(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.GetArchiveByID")

	archiveID := c.Param("archive_id")
	characterVersion := c.Param("character_version")
	profileID := c.Param("profile_id")

	a, err := characters.GetArchiveByID(ctx, logCtx, archiveID, profileID, characterVersion)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get character archive")
	}

	return c.JSON(http.StatusOK, a)
}

// GetSessionEntryByID returns a session entry by ID.
func GetSessionEntryByID(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.GetSessionEntryByID")

	sessionID := c.Param("session_id")
	characterVersion := c.Param("character_version")
	profileID := c.Param("profile_id")

	sessID, err := strconv.Atoi(sessionID)
	if err != nil {
		return e.ErrBad(logCtx, fid, "invalid session_id")
	}

	a, err := characters.GetSessionEntryByID(ctx, logCtx, profileID, characterVersion, sessID)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get character session")
	}

	return c.JSON(http.StatusOK, a)
}

// GetArchiveIndex returns the profile archive index.
func GetArchiveIndex(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.GetArchiveIndex")

	characterVersion := c.Param("character_version")
	profileID := c.Param("profile_id")

	a, err := characters.GetArchiveIndex(ctx, logCtx, profileID, characterVersion)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get character archives")
	}

	return c.JSON(http.StatusOK, a)
}

// GetArchiveEntriesByDateRange returns a profile archive summary by date.
func GetArchiveEntriesByDateRange(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.GetArchiveEntriesByDateRange")

	characterVersion := c.Param("character_version")
	profileID := c.Param("profile_id")
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	start, err := time.Parse(time.RFC3339, startDate)
	if err != nil {
		return e.ErrBad(logCtx, fid, "start date required (expected RFC3339)")
	}

	var end time.Time
	if endDate == "" {
		end = start.Add(24 * time.Hour)
	} else {
		end, err = time.Parse(time.RFC3339, endDate)
		if err != nil {
			return e.ErrBad(logCtx, fid, "end date bad format (expected RFC3339)")
		}
	}

	if start.After(end) {
		return e.ErrBad(logCtx, fid, "start date cannot be after the end date")
	}

	a, err := characters.GetArchiveEntriesByDateRange(ctx, logCtx, profileID, characterVersion, start, end)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get character archive entries")
	}

	return c.JSON(http.StatusOK, a)
}

// GetArchiveSummaryByDateRange returns a profile archive summary by date.
func GetArchiveSummaryByDateRange(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.GetArchiveSummaryByDateRange")

	characterVersion := c.Param("character_version")
	profileID := c.Param("profile_id")
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	start, err := time.Parse(time.RFC3339, startDate)
	if err != nil {
		return e.ErrBad(logCtx, fid, "archive start date required (expected RFC3339)")
	}

	var end time.Time
	if endDate == "" {
		end = start.Add(24 * time.Hour)
	} else {
		end, err = time.Parse(time.RFC3339, endDate)
		if err != nil {
			return e.ErrBad(logCtx, fid, "end date bad format (expected RFC3339)")
		}
	}

	if start.After(end) {
		return e.ErrBad(logCtx, fid, "start date cannot be after the end date")
	}

	a, err := characters.GetArchiveSummaryByDateRange(ctx, logCtx, profileID, characterVersion, start, end)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get character archive summary")
	}

	return c.JSON(http.StatusOK, a)
}

// DeleteArchiveSummaryDateRange deletes a profile's archive summary.
func DeleteArchiveSummaryDateRange(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.DeleteArchiveSummaryDateRange")

	characterVersion := c.Param("character_version")
	profileID := c.Param("profile_id")

	if err := characters.DeleteArchiveSummaryDateRange(ctx, logCtx, profileID, characterVersion); err != nil {
		return e.Err(logCtx, err, fid, "unable to delete character archive date range")
	}

	return c.NoContent(http.StatusNoContent)
}

// DeleteSessionMemory deletes a profile's memory.
func DeleteSessionMemory(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.profiles.DeleteSessionMemory")

	characterVersion := c.Param("character_version")
	profileID := c.Param("profile_id")

	if err := characters.DeleteSessionMemory(ctx, logCtx, profileID, characterVersion); err != nil {
		return e.Err(logCtx, err, fid, "unable to delete character session memory")
	}

	return c.NoContent(http.StatusNoContent)
}
