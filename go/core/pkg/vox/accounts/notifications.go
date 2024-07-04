package accounts

import (
	"context"
	"log/slog"

	"firebase.google.com/go/messaging"

	"disruptive/lib/firebase"
)

// PushNotificationRequest is a request structure to send a notification.
type PushNotificationRequest struct {
	FCMToken string `json:"fcm_token"`
	Title    string `json:"title"`
	Body     string `json:"body"`
}

// PushNotification sends a Firebase notification to a device.
func PushNotification(ctx context.Context, logCtx *slog.Logger, req PushNotificationRequest) (string, error) {
	fid := slog.String("fid", "vox.accounts.PushNotification")

	message := &messaging.Message{
		Token: req.FCMToken,
		Notification: &messaging.Notification{
			Title: req.Title,
			Body:  req.Body,
		},
	}

	client, err := firebase.App.Messaging(ctx)
	if err != nil {
		logCtx.Error("unable to get firebase messaging client", fid, "error", err)
		return "", err
	}

	res, err := client.Send(ctx, message)
	if err != nil {
		logCtx.Error("unable to send message", fid, "error", err)
		return "", err
	}

	return res, nil
}
