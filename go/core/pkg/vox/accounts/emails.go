package accounts

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"disruptive/config"
)

// EmailRequest is a request structure to send an email.
type EmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
	HTML    string   `json:"html"`
}

// SendEmail sends an email.
func SendEmail(ctx context.Context, logCtx *slog.Logger, req EmailRequest) error {
	fid := slog.String("fid", "vox.accounts.SendEmail")

	data := url.Values{}
	data.Add("from", req.From)

	for _, t := range req.To {
		data.Add("to", t)
	}
	data.Add("subject", req.Subject)

	if req.Text != "" {
		data.Add("text", req.Text)
	}

	if req.HTML != "" {
		data.Add("html", req.HTML)
	}

	res, err := Resty.R().
		SetContext(ctx).
		SetPathParam("domain", config.VARS.MailgunDomain).
		SetFormDataFromValues(data).
		Post(messagesEndpoint)

	if err != nil {
		logCtx.Error("mailgun endpoint failed", fid, "error", err)
		return err
	}

	if res.StatusCode() != http.StatusOK {
		logCtx.Error("mailgun endpoint failed", fid, "status", res.Status())
		return errors.New(res.Status())
	}

	return nil
}
