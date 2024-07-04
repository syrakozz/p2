package accounts

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/pkg/vox/accounts"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// PostPushNotifications sends a notification to a device
func PostPushNotifications(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.PostPushNotifications")

	req := accounts.PushNotificationRequest{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	res, err := accounts.PushNotification(ctx, logCtx, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unabel to send notification")
	}

	return c.JSON(http.StatusCreated, map[string]string{"response": res})
}
