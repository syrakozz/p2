package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"disruptive/lib/common"
	"disruptive/lib/firestore"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/characters"

	"github.com/google/uuid"
)

// ModerationValue contains the moderation notification information.
type ModerationValue struct {
	Profile   ModerationProfileValue   `firestore:"profile" json:"profile"`
	Character ModerationCharacterValue `firestore:"character" json:"character"`
	Session   *ModerationSessionValue  `firestore:"session,omitempty" json:"session,omitempty"`
}

// ModerationProfileValue ...
type ModerationProfileValue struct {
	ID   string `firestore:"id" json:"id"`
	Name string `firestore:"name" json:"name"`
}

// ModerationCharacterValue ...
type ModerationCharacterValue struct {
	Name string `firestore:"name" json:"name"`
}

// ModerationSessionValue ...
type ModerationSessionValue struct {
	Archive     time.Time               `firestore:"archive" json:"archive"`
	Entry       characters.SessionEntry `firestore:"entry" json:"entry"`
	EntryNumber int                     `firestore:"entry_number" json:"entry_number"`
}

// PostModeration creates a new notification
func PostModeration(ctx context.Context, logCtx *slog.Logger, req ModerationValue) (Document, error) {
	fid := slog.String("fid", "vox.notifications.PostModeration")

	account := ctx.Value(common.AccountKey).(accounts.Document)

	path := fmt.Sprintf("accounts/%s/notifications", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("notifications collection not found", fid)
		return Document{}, common.ErrNotFound{}
	}

	uuid := uuid.New().String()

	document := Document{
		ID:              uuid,
		Type:            "moderation",
		Timestamp:       time.Now(),
		ModerationValue: &req,
	}

	if _, err := collection.Doc(uuid).Create(ctx, document); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to create notification document", fid, "error", err)
		return Document{}, err
	}

	return document, nil
}
