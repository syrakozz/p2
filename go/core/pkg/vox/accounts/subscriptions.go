package accounts

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	fs "cloud.google.com/go/firestore"

	"disruptive/lib/common"
	"disruptive/lib/configs"
	"disruptive/lib/firestore"
)

// Subscription contains subscription information
type Subscription struct {
	SubscriptionBalance   int       `firestore:"subscription_balance" json:"subscription_balance"`
	SubscriptionPending   string    `firestore:"subscription_pending" json:"subscription_pending,omitempty"`
	SubscriptionSKU       string    `firestore:"subscription_sku" json:"subscription_sku"`
	SubscriptionStartDate time.Time `firestore:"subscription_start_date" json:"subscription_start_date"`
}

// GetSubscription gets a user's subscription information.
func GetSubscription(ctx context.Context, logCtx *slog.Logger) (Subscription, error) {
	fid := slog.String("fid", "vox.accounts.GetSubscription")

	b, err := GetBalance(ctx, logCtx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get account balance document", fid, "error", err)
		return Subscription{}, err
	}

	if time.Now().After(b.SubscriptionStartDate) && b.SubscriptionPending != "" {
		b, err = Subscribe(ctx, logCtx, b.SubscriptionPending)
		if err != nil {
			logCtx.Error("unable to update subscription", fid)
			return Subscription{}, err
		}
	}

	s := Subscription{
		SubscriptionBalance:   b.SubscriptionBalance,
		SubscriptionPending:   b.SubscriptionPending,
		SubscriptionSKU:       b.SubscriptionSKU,
		SubscriptionStartDate: b.SubscriptionStartDate,
	}

	return s, nil
}

// PatchPendingSubscription sets the subscription pending value.
func PatchPendingSubscription(ctx context.Context, logCtx *slog.Logger, sku string) (BalanceDocument, error) {
	fid := slog.String("fid", "vox.accounts.PatchPendingSubscription")

	account := ctx.Value(common.AccountKey).(Document)

	skus, err := configs.GetSKUs(ctx, logCtx)
	if err != nil {
		logCtx.Error("unable to get banking configs", fid)
		return BalanceDocument{}, common.ErrNotFound{}
	}

	_, ok := skus[sku]
	if !ok {
		logCtx.Error("invalid sku", fid)
		return BalanceDocument{}, common.ErrBadRequest{Src: "Subscribe", Msg: "invalid sku"}
	}

	path := fmt.Sprintf("accounts/%s/bank", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return BalanceDocument{}, common.ErrNotFound{}
	}

	updates := []fs.Update{
		{Path: "subscription_pending", Value: sku},
	}

	if len(updates) > 0 {
		if _, err := collection.Doc("balance").Update(ctx, updates); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Warn("unable to update accounts bank document", fid, "error", err)
			return BalanceDocument{}, err
		}
	}

	b, err := GetBalance(ctx, logCtx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get account balance document", fid, "error", err)
		return BalanceDocument{}, err
	}

	return b, nil
}

// Subscribe updates an accounts subscription.
func Subscribe(ctx context.Context, logCtx *slog.Logger, sku string) (BalanceDocument, error) {
	fid := slog.String("fid", "vox.accounts.Subscribe")

	account := ctx.Value(common.AccountKey).(Document)

	b, err := GetBalance(ctx, logCtx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get account balance document", fid, "error", err)
		return BalanceDocument{}, err
	}

	skus, err := configs.GetSKUs(ctx, logCtx)
	if err != nil {
		logCtx.Error("unable to get banking configs", fid)
		return BalanceDocument{}, common.ErrNotFound{}
	}

	newSub, ok := skus[sku]
	if !ok {
		logCtx.Error("invalid sku", fid)
		return BalanceDocument{}, common.ErrBadRequest{Src: "Subscribe", Msg: "invalid sku"}
	}

	path := fmt.Sprintf("accounts/%s/bank", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return BalanceDocument{}, common.ErrNotFound{}
	}

	updates := []fs.Update{
		{Path: "subscription_balance", Value: newSub.Balance},
		{Path: "subscription_sku", Value: sku},
		{Path: "subscription_pending", Value: fs.Delete},
		{Path: "subscription_start_date", Value: b.SubscriptionStartDate.AddDate(0, 1, 0)},
	}

	if len(updates) > 0 {
		if _, err := collection.Doc("balance").Update(ctx, updates); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Warn("unable to update accounts bank document", fid, "error", err)
			return BalanceDocument{}, err
		}
	}

	b, err = GetBalance(ctx, logCtx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get account balance document", fid, "error", err)
		return BalanceDocument{}, err
	}

	return b, nil
}

// Unsubscribe unsubscribes a user.
func Unsubscribe(ctx context.Context, logCtx *slog.Logger) (BalanceDocument, error) {
	fid := slog.String("fid", "vox.accounts.Subscribe")

	account := ctx.Value(common.AccountKey).(Document)

	path := fmt.Sprintf("accounts/%s/bank", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return BalanceDocument{}, common.ErrNotFound{}
	}

	updates := []fs.Update{
		{Path: "subscription_sku", Value: fs.Delete},
	}

	if _, err := collection.Doc("balance").Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update accounts bank document", fid, "error", err)
		return BalanceDocument{}, err
	}

	b, err := GetBalance(ctx, logCtx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get account balance document", fid, "error", err)
		return BalanceDocument{}, err
	}

	return b, nil
}
