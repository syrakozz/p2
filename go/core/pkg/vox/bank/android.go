package bank

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	fs "cloud.google.com/go/firestore"

	"disruptive/lib/common"
	"disruptive/lib/configs"
	"disruptive/lib/firestore"
)

// GoogleAndroidIAPTransaction contains an android iap transaction. Inbound only.
type GoogleAndroidIAPTransaction struct {
	Acknowledged  bool   `json:"acknowledged"`
	OrderID       string `json:"orderId"`
	PackageName   string `json:"packageName"`
	ProductID     string `json:"productId"`
	PurchaseState int    `json:"purchaseState"`
	PurchaseTime  int    `json:"purchaseTime"`
	PurchaseToken string `json:"purchaseToken"`
	Quantity      int    `json:"quantity"`
}

// AndroidIAPTransaction contains an android iap transaction. DB and outbound only.
type AndroidIAPTransaction struct {
	Acknowledged  bool      `firestore:"acknowledged,omitempty" json:"acknowledged,omitempty"`
	Amount        float64   `firestore:"amount" json:"amount"`
	OrderID       string    `firestore:"order_id,omitempty" json:"order_id,omitempty"`
	PackageName   string    `firestore:"package_name" json:"package_name"`
	ProductID     string    `firestore:"product_id" json:"product_id"`
	PurchaseState int       `firestore:"purchase_state,omitempty" json:"purchase_state,omitempty"`
	PurchaseTime  time.Time `firestore:"purchase_time,omitempty" json:"purchase_time,omitempty"`
	PurchaseToken string    `firestore:"purchase_token" json:"purchase_token"`
	Quantity      int       `firestore:"quantity" json:"quantity"`
}

// AndroidSubTransaction contains a android subscription transaction.
type AndroidSubTransaction struct {
	AccountID      string    `firestore:"account_id"`
	CountryCode    string    `firestore:"country_code" json:"country_code"`
	LoadBalancer   string    `firestore:"load_balancer" json:"load_balancer"`
	PurchaseToken  string    `firestore:"purchase_token" json:"purchase_token"` // document_id = purchase_token
	SKU            string    `firestore:"sku" json:"sku"`
	SubscriptionID string    `firestore:"subscription_id" json:"subscription_id"` // user's order number
	Timestamp      time.Time `firestore:"timestamp"`
}

func renderAndroidIAPTransaction(txn GoogleAndroidIAPTransaction) AndroidIAPTransaction {
	_txn := AndroidIAPTransaction{
		OrderID:       txn.OrderID,
		PackageName:   txn.PackageName,
		ProductID:     txn.ProductID,
		PurchaseState: txn.PurchaseState,
		PurchaseToken: txn.PurchaseToken,
		Quantity:      txn.Quantity,
		Acknowledged:  txn.Acknowledged,
	}

	if txn.PurchaseTime != 0 {
		t := int64(txn.PurchaseTime)
		s := t / 1000
		ns := (t & 1000) * 1000000

		_txn.PurchaseTime = time.Unix(s, ns).UTC()
	}

	return _txn
}

// PostAndroidIAPTransaction creates a new android transaction.
func PostAndroidIAPTransaction(ctx context.Context, logCtx *slog.Logger, uid string, txn GoogleAndroidIAPTransaction) (AndroidIAPTransaction, error) {
	fid := slog.String("fid", "vox.bank.PostAndroidIAPTransaction")

	_txn := renderAndroidIAPTransaction(txn)

	collection := firestore.Client.Collection(fmt.Sprintf("accounts/%s/android_iap_transactions", uid))
	if collection == nil {
		logCtx.Error("android IAP transactions collection not found", fid)
		return AndroidIAPTransaction{}, common.ErrNotFound{}
	}

	doc, err := collection.Doc(_txn.PurchaseToken).Get(ctx)
	if err != nil && !errors.Is(common.ConvertGRPCError(err), common.ErrNotFound{}) {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get android IAP transactions document", fid, "error", err)
		return AndroidIAPTransaction{}, err
	}

	if doc.Exists() {
		logCtx.Warn("transaction already exists", fid)
		return AndroidIAPTransaction{}, common.ErrAlreadyExists{}
	}

	skus, err := configs.Get(ctx, logCtx, "skus")
	if err != nil {
		logCtx.Error("unable to get sku config", fid, "error", err)
		return AndroidIAPTransaction{}, common.ErrNotFound{}
	}

	sku, ok := skus[_txn.ProductID].(map[string]any)
	if !ok {
		logCtx.Error("sku not found", fid)
		return AndroidIAPTransaction{}, common.ErrNotFound{}
	}

	switch b := sku["balance"].(type) {
	case int64:
		_txn.Amount = float64(b)
	case float64:
		_txn.Amount = b
	default:
		logCtx.Error("balance for sku not found", fid, "sku", sku)
		return AndroidIAPTransaction{}, common.ErrNotFound{}
	}

	if _, err := collection.Doc(_txn.PurchaseToken).Set(ctx, _txn); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to create android IAP transactions document", fid, "error", err)
		return AndroidIAPTransaction{}, err
	}

	collection = firestore.Client.Collection(fmt.Sprintf("accounts/%s/bank", uid))
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return AndroidIAPTransaction{}, common.ErrNotFound{}
	}

	updates := []fs.Update{{Path: "balance", Value: fs.Increment(_txn.Amount * float64(_txn.Quantity))}}

	if _, err := collection.Doc("balance").Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update accounts bank document", fid, "error", err)
		return AndroidIAPTransaction{}, err
	}

	return _txn, nil
}

// PostAndroidSubTransaction creates a new android transaction.
func PostAndroidSubTransaction(ctx context.Context, logCtx *slog.Logger, txn AndroidSubTransaction) (AndroidSubTransaction, error) {
	fid := slog.String("fid", "vox.bank.PostAndroidSubTransaction")

	collection := firestore.Client.Collection("stores/android/transactions")
	if collection == nil {
		logCtx.Error("memory collection not found", fid)
		return AndroidSubTransaction{}, common.ErrNotFound{}
	}

	if _, err := collection.Doc(txn.PurchaseToken).Set(ctx, txn); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update android transactions document", fid, "error", err)
		return AndroidSubTransaction{}, err
	}

	doc, err := collection.Doc(txn.PurchaseToken).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get android transactions document", fid, "error", err)
		return AndroidSubTransaction{}, err
	}

	t := AndroidSubTransaction{}

	if err := doc.DataTo(&t); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read data", fid, "error", err)
		return AndroidSubTransaction{}, err
	}

	return t, nil
}
