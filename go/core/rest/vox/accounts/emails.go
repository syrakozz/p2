package accounts

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/notifications"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// PostEmails sends email.
func PostEmails(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.PostEmails")

	req := accounts.EmailRequest{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read data")
	}

	if req.From == "" {
		return e.ErrBad(logCtx, fid, "invalid email from")
	}

	if len(req.To) == 0 {
		return e.ErrBad(logCtx, fid, "invalid email to")
	}

	if req.Subject == "" {
		return e.ErrBad(logCtx, fid, "invalid email subject")
	}

	if req.HTML == "" && req.Text == "" {
		return e.ErrBad(logCtx, fid, "invalid email body")
	}

	if err := accounts.SendEmail(ctx, logCtx, req); err != nil {
		return e.Err(logCtx, err, fid, "unable to send email")
	}

	return c.NoContent(http.StatusCreated)
}

// PostEmailPin sends the current account pin to the account email address.
func PostEmailPin(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.accounts.PostEmailPin")

	language := c.QueryParam("language")

	switch language {
	case "", "en":
		language = "en-US"
	case "es":
		language = "es-419"
	case "fr":
		language = "fr-FR"
	case "pt":
		language = "pt-PT"
	}

	if err := notifications.SendPinEmail(ctx, logCtx, language); err != nil {
		return e.Err(logCtx, err, fid, "unable to send email")
	}

	return c.NoContent(http.StatusNoContent)
}
