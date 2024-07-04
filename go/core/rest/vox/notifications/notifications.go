package notifications

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/pkg/vox/notifications"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// GetAll retrieves all account notifications.
func GetAll(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.notifications.GetAll")

	all := c.QueryParam("all") == "true"
	inactive := c.QueryParam("inactive") == "true"

	if all && inactive {
		return e.ErrBad(logCtx, fid, "all and inactive cannot both be true")
	}

	n, err := notifications.GetNotifications(ctx, logCtx, all, inactive)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get notification")
	}

	return c.JSON(http.StatusOK, n)
}

// Post creates an account notification.
func Post(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.notifications.Post")

	nType := c.QueryParam("type")
	logCtx = logCtx.With("type", nType)

	if nType == "moderation" {
		v := notifications.ModerationValue{}

		if err := c.Bind(&v); err != nil {
			return e.ErrBad(logCtx, fid, "unable to read data")
		}

		n, err := notifications.PostModeration(ctx, logCtx, v)
		if err != nil {
			return e.Err(logCtx, err, fid, "unable to create moderation notification")
		}

		return c.JSON(http.StatusOK, n)
	}

	return e.ErrBad(logCtx, fid, "invalid notification type")
}

// Get retrieves an account notification.
func Get(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.notifications.Get")

	id := c.Param("notification_id")

	n, err := notifications.GetNotification(ctx, logCtx, id)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get notification")
	}

	return c.JSON(http.StatusOK, n)

}

// Patch modifies an account notification.
func Patch(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.notifications.Patch")

	id := c.Param("notification_id")

	p := notifications.PatchDocument{}

	if err := c.Bind(&p); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	p.ID = id

	n, err := notifications.PatchNotification(ctx, logCtx, p)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to patch notification")
	}

	return c.JSON(http.StatusOK, n)
}

// Delete deletes an account notification.
func Delete(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.notifications.Delete")

	id := c.Param("notification_id")

	if err := notifications.DeleteNotification(ctx, logCtx, id); err != nil {
		return e.Err(logCtx, err, fid, "unable to delete notification")
	}

	return c.NoContent(http.StatusNoContent)
}
