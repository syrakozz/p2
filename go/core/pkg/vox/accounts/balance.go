package accounts

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	fs "cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"disruptive/lib/common"
	"disruptive/lib/configs"
	"disruptive/lib/firestore"
	"disruptive/pkg/vox/bank"
)

// BalanceDocument is the balance document
type BalanceDocument struct {
	Balance                int       `firestore:"balance" json:"balance"`
	FirstTimeIOSFirmware20 bool      `firestore:"first_time_ios_firmware_20" json:"first_time_ios_firmware_20,omitempty"`
	SubscriptionBalance    int       `firestore:"subscription_balance" json:"subscription_balance"`
	SubscriptionPending    string    `firestore:"subscription_pending" json:"subscription_pending,omitempty"`
	SubscriptionSKU        string    `firestore:"subscription_sku" json:"subscription_sku"`
	SubscriptionStartDate  time.Time `firestore:"subscription_start_date" json:"subscription_start_date"`
}

// BalanceInfo contains an accounts balances
type BalanceInfo struct {
	Balance             int `json:"balance"`
	CodesBalance        int `json:"codes_balance,omitempty"`
	SubscriptionBalance int `json:"subscription_balance"`
	TotalBalance        int `json:"total_balance"`
}

// RedeemedCardInfo contains info on redeeming a gift card
type RedeemedCardInfo struct {
	Balance       int    `json:"balance"`
	Code          string `json:"code"`
	ValueRedeemed int    `json:"value_redeemed"`
}

// GetBalance retrieves a firestore account's bank balance info.
func GetBalance(ctx context.Context, logCtx *slog.Logger) (BalanceDocument, error) {
	fid := slog.String("fid", "vox.accounts.GetBank")

	account := ctx.Value(common.AccountKey).(Document)

	path := fmt.Sprintf("accounts/%s/bank", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("unable to get account collection", fid)
		return BalanceDocument{}, common.ErrNotFound{}
	}

	doc, err := collection.Doc("balance").Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			if err := CreateBalanceDocument(ctx, logCtx, ""); err != nil {
				err = common.ConvertGRPCError(err)
				logCtx.Error("unable to create accounts balance document", fid, "error", err)
				return BalanceDocument{}, err
			}
			return BalanceDocument{}, nil
		}

		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get accounts balance document", fid, "error", err)
		return BalanceDocument{}, err
	}

	d := BalanceDocument{}

	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read account bank data", fid, "error", err)
		return BalanceDocument{}, err
	}

	return d, nil
}

// GetAvailableBalance retrieves a firestore account's balances.
func GetAvailableBalance(ctx context.Context, logCtx *slog.Logger) (BalanceInfo, error) {
	fid := slog.String("fid", "vox.accounts.GetAvailableBalance")

	account := ctx.Value(common.AccountKey).(Document)

	path := fmt.Sprintf("accounts/%s/bank", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("unable to get account collection", fid)
		return BalanceInfo{}, common.ErrNotFound{}
	}

	doc, err := collection.Doc("balance").Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			if err := CreateBalanceDocument(ctx, logCtx, ""); err != nil {
				err = common.ConvertGRPCError(err)
				logCtx.Error("unable to create accounts balance document", fid, "error", err)
				return BalanceInfo{}, err
			}
			return BalanceInfo{}, nil
		}

		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get accounts balance document", fid, "error", err)
		return BalanceInfo{}, err
	}

	d := BalanceDocument{}

	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read account bank data", fid, "error", err)
		return BalanceInfo{}, err
	}

	b := BalanceInfo{
		Balance:             d.Balance,
		SubscriptionBalance: d.SubscriptionBalance,
	}

	b.TotalBalance = d.Balance + d.SubscriptionBalance

	return b, nil
}

// AddBalance adds the amount to the user's balance.
func AddBalance(ctx context.Context, logCtx *slog.Logger, amount int) (BalanceDocument, error) {
	fid := slog.String("fid", "vox.accounts.AddBalance")

	account := ctx.Value(common.AccountKey).(Document)

	path := fmt.Sprintf("accounts/%s/bank", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return BalanceDocument{}, common.ErrNotFound{}
	}

	updates := []fs.Update{
		{Path: "balance", Value: fs.Increment(amount)},
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
		logCtx.Error("unable to get account balance document", fid, "error", err)
		return BalanceDocument{}, err
	}

	return b, nil
}

// ChargeBank charges an account's bank.
func ChargeBank(ctx context.Context, logCtx *slog.Logger, characterVersion, tier string) (BalanceDocument, error) {
	fid := slog.String("fid", "vox.accounts.ChargeBank")

	account := ctx.Value(common.AccountKey).(Document)

	path := fmt.Sprintf("accounts/%s/bank", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return BalanceDocument{}, common.ErrNotFound{}
	}

	rates, err := configs.GetRates(ctx, logCtx)
	if err != nil {
		logCtx.Error("unable to get rates", fid, "error", err)
		return BalanceDocument{}, err
	}

	cost, ok := rates[tier]
	if !ok {
		logCtx.Error("unable to get character tier and rate", fid, "error", err, "character", characterVersion)
		return BalanceDocument{}, common.ErrNotFound{Msg: "unable to get character tier and rate"}
	}

	updates := []fs.Update{}

	balance, err := GetAvailableBalance(ctx, logCtx)
	if err != nil {
		logCtx.Error("unable to get bank balance", fid, "error", err)
		return BalanceDocument{}, err
	}

	// have enough balance to pay outright
	if balance.SubscriptionBalance >= cost {
		updates = append(updates, fs.Update{Path: "subscription_balance", Value: fs.Increment(-cost)})
	}

	if len(updates) == 0 {
		if balance.Balance >= cost {
			updates = append(updates, fs.Update{Path: "balance", Value: fs.Increment(-cost)})
		}
	}

	// need to split cost across both balances
	if len(updates) == 0 {
		totalPaid := 0
		if balance.TotalBalance > 0 {
			if balance.SubscriptionBalance > 0 {
				updates = append(updates, fs.Update{Path: "subscription_balance", Value: 0})
				totalPaid += balance.SubscriptionBalance
			}

			remainingPayment := cost - totalPaid
			if balance.Balance > 0 && remainingPayment > 0 {
				if balance.Balance >= remainingPayment {
					updates = append(updates, fs.Update{Path: "balance", Value: fs.Increment(-remainingPayment)})
				} else {
					updates = append(updates, fs.Update{Path: "balance", Value: 0})
				}
			}
		}
	}

	if len(updates) == 0 {
		return BalanceDocument{}, common.ErrPaymentRequired{Src: "ChargeBank", Msg: "user does not have sufficient balance"}
	}

	if _, err := collection.Doc("balance").Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update account bank", fid, "error", err)
		return BalanceDocument{}, err
	}

	b, err := GetBalance(ctx, logCtx)
	if err != nil {
		logCtx.Error("unable to get account balance document", fid, "error", err)
		return BalanceDocument{}, err
	}

	return b, nil
}

// RedeemGiftCard redeems a gift card and adds the balance to the account.
func RedeemGiftCard(ctx context.Context, logCtx *slog.Logger, giftCard string) (RedeemedCardInfo, error) {
	fid := slog.String("fid", "vox.accounts.RedeemGiftCard")

	account := ctx.Value(common.AccountKey).(Document)

	gcs, err := bank.GetGiftCards(ctx, logCtx, "")
	if err != nil {
		logCtx.Error("unable to get current gift cards", fid, "error", err)
		return RedeemedCardInfo{}, err
	}

	gc, ok := gcs[giftCard]
	if !ok {
		logCtx.Error("invalid gift card", fid)
		return RedeemedCardInfo{}, common.ErrNotFound{Msg: "invalid gift card"}
	}

	if !gc.Start.IsZero() && gc.Start.After(time.Now()) {
		logCtx.Warn("gift card not started", fid)
		return RedeemedCardInfo{}, common.ErrUnprocessable{Msg: "gift card not started"}
	}

	path := fmt.Sprintf("accounts/%s/bank", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return RedeemedCardInfo{}, common.ErrNotFound{}
	}

	updates := []fs.Update{
		{Path: "balance", Value: fs.Increment(gc.Value)},
	}

	if len(updates) > 0 {
		if _, err := collection.Doc("balance").Update(ctx, updates); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Warn("unable to update promo codes", fid, "error", err)
			return RedeemedCardInfo{}, err
		}
	}

	if gc.OneTimeUse {
		if err := bank.ExpireGiftCard(ctx, logCtx, giftCard); err != nil {
			logCtx.Error("unable to expire one time use gift card", fid)
			return RedeemedCardInfo{}, err
		}
	}

	b, err := GetBalance(ctx, logCtx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get account balance document", fid, "error", err)
		return RedeemedCardInfo{}, err
	}

	return RedeemedCardInfo{
		Balance:       b.Balance,
		Code:          giftCard,
		ValueRedeemed: int(gc.Value),
	}, nil
}

// CreateBalanceDocument creates an account's balance document.
func CreateBalanceDocument(ctx context.Context, logCtx *slog.Logger, accountID string) error {
	fid := slog.String("fid", "vox.accounts.CreateBalanceDocument")

	path := fmt.Sprintf("accounts/%s/bank", accountID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("unable to get account collection", fid)
		return common.ErrNotFound{}
	}

	_, err := collection.Doc("balance").Set(ctx, BalanceDocument{})
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get create accounts balance document", fid, "error", err)
		return err
	}

	return nil
}

// BumpTo20kVexels TO BE REMOVED: Quick fix. Put balances <=10,500 back up to 20,000.
func BumpTo20kVexels(ctx context.Context, logCtx *slog.Logger, balance int) error {
	fid := slog.String("fid", "vox.accounts.BumpTo20kVexels")

	valueTo20k := 20000 - balance

	// bump balance to 20k vexels
	if _, err := AddBalance(ctx, logCtx, valueTo20k); err != nil {
		logCtx.Error("unable to add to account balance", fid)
		return err
	}

	// update account used_codes with a timestamp of the free vexel transaction
	account := ctx.Value(common.AccountKey).(Document)

	collection := firestore.Client.Collection(fmt.Sprintf("accounts/%s/bank", account.ID))
	if collection == nil {
		logCtx.Error("unable to get account's bank collection", fid)
		return common.ErrNotFound{}
	}

	updates := []fs.Update{
		{Path: fmt.Sprintf("free_vexels_%s.vexels", time.Now().Format(time.RFC3339)), Value: valueTo20k},
		{Path: "free_vexels_added_total.vexels", Value: fs.Increment(valueTo20k)},
		{Path: "free_vexels_added_total.num_times", Value: fs.Increment(1)},
	}

	if _, err := collection.Doc("used_codes").Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update account document", fid, "error", err)
		return err
	}

	// update bank tally with how many free vexels have been added so far.
	// storing in bank's gift_card document fields just as an easy place holder to save this data.
	collection = firestore.Client.Collection("bank")
	if collection == nil {
		logCtx.Error("unable to get bank collection", fid)
		return common.ErrNotFound{}
	}

	updates = []fs.Update{
		{Path: "total_free_vexels", Value: fs.Increment(valueTo20k)},
		{Path: "total_num_times", Value: fs.Increment(1)},
	}

	if _, err := collection.Doc("gift_cards").Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update account document", fid, "error", err)
		return err
	}

	return nil
}
