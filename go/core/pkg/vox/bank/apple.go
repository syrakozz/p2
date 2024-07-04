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

// AppleStoreIAPTransaction contains an apple iap transaction from the store. Inbound only.
type AppleStoreIAPTransaction struct {
	ProductID          string `json:"productId"`
	TransactionID      string `json:"transactionId"`
	TransactionDate    int    `json:"transactionDate"`
	TransactionReceipt string `json:"transactionReceipt"`
}

// AppleIAPTransaction contains an apple iap transaction. DB and outbound only.
type AppleIAPTransaction struct {
	Amount             float64   `firestore:"amount" json:"amount"`
	ProductID          string    `firestore:"product_id" json:"product_id"`
	TransactionID      string    `firestore:"transaction_id" json:"transaction_id"`
	TransactionDate    time.Time `firestore:"transaction_date,omitempty" json:"transaction_date,omitempty"`
	TransactionReceipt string    `firestore:"transaction_receipt" json:"transaction_receipt"`
}

func renderAppleIAPTransaction(txn AppleStoreIAPTransaction) AppleIAPTransaction {
	_txn := AppleIAPTransaction{
		ProductID:          txn.ProductID,
		TransactionID:      txn.TransactionID,
		TransactionReceipt: txn.TransactionReceipt,
	}

	if txn.TransactionDate != 0 {
		t := int64(txn.TransactionDate)
		s := t / 1000
		ns := (t & 1000) * 1000000

		_txn.TransactionDate = time.Unix(s, ns).UTC()
	}

	return _txn
}

// PostAppleIAPTransaction creates a new apple iap transaction.
func PostAppleIAPTransaction(ctx context.Context, logCtx *slog.Logger, uid string, txn AppleStoreIAPTransaction) (AppleIAPTransaction, error) {
	fid := slog.String("fid", "vox.bank.PostAppleIAPTransaction")

	_txn := renderAppleIAPTransaction(txn)

	collection := firestore.Client.Collection(fmt.Sprintf("accounts/%s/apple_iap_transactions", uid))
	if collection == nil {
		logCtx.Error("apple IAP transactions collection not found", fid)
		return AppleIAPTransaction{}, common.ErrNotFound{}
	}

	doc, err := collection.Doc(_txn.TransactionID).Get(ctx)
	if err != nil && !errors.Is(common.ConvertGRPCError(err), common.ErrNotFound{}) {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get apple IAP transactions document", fid, "error", err)
		return AppleIAPTransaction{}, err
	}

	if doc.Exists() {
		logCtx.Warn("transaction already exists", fid)
		return AppleIAPTransaction{}, common.ErrAlreadyExists{}
	}

	skus, err := configs.Get(ctx, logCtx, "skus")
	if err != nil {
		logCtx.Error("unable to get sku config", fid, "error", err)
		return AppleIAPTransaction{}, common.ErrNotFound{}
	}

	sku, ok := skus[_txn.ProductID].(map[string]any)
	if !ok {
		logCtx.Error("sku not found", fid)
		return AppleIAPTransaction{}, common.ErrNotFound{}
	}

	switch b := sku["balance"].(type) {
	case int64:
		_txn.Amount = float64(b)
	case float64:
		_txn.Amount = b
	default:
		logCtx.Error("balance for sku not found", fid, "sku", sku)
		return AppleIAPTransaction{}, common.ErrNotFound{}
	}

	if _, err := collection.Doc(_txn.TransactionID).Set(ctx, _txn); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to create apple IAP transactions document", fid, "error", err)
		return AppleIAPTransaction{}, err
	}

	collection = firestore.Client.Collection(fmt.Sprintf("accounts/%s/bank", uid))
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return AppleIAPTransaction{}, common.ErrNotFound{}
	}

	updates := []fs.Update{{Path: "balance", Value: fs.Increment(_txn.Amount)}}

	if _, err := collection.Doc("balance").Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update accounts bank document", fid, "error", err)
		return AppleIAPTransaction{}, err
	}

	return _txn, nil
}
