package configs

import (
	"context"
	"disruptive/lib/common"
	"disruptive/lib/firestore"
	"errors"
	"log/slog"

	fs "cloud.google.com/go/firestore"
)

// Rates are the character rates and tiers
type Rates map[string]int

// SKU defines a specific subscription model
type SKU struct {
	Balance     int    `firestore:"balance" json:"balance"`
	Description string `firestore:"description" json:"description"`
	Period      string `firestore:"period" json:"period"`
	Rollover    bool   `firestore:"rollover" json:"rollover"`
	Title       string `firestore:"title" json:"title"`
}

// SKUs are the list of available skus and their info
type SKUs map[string]SKU

// GetSKUs returns the banking SKUs.
func GetSKUs(ctx context.Context, logCtx *slog.Logger) (SKUs, error) {
	fid := slog.String("fid", "console.configs.GetSKUs")

	collection := firestore.Client.Collection("configs")
	if collection == nil {
		logCtx.Warn("configs collection not found", fid)
		return SKUs{}, common.ErrNotFound{}
	}

	var (
		doc *fs.DocumentSnapshot
		err error
	)

	doc, err = collection.Doc("skus").Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		if !errors.Is(err, common.ErrNotFound{}) {
			logCtx.Error("unable to get bank config", fid, "error", err)
			return SKUs{}, err
		}
	}

	s := SKUs{}

	if err := doc.DataTo(&s); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read configs SKU data", fid, "error", err)
		return map[string]SKU{}, err
	}

	return s, nil
}

// GetRates returns the banking character rates.
func GetRates(ctx context.Context, logCtx *slog.Logger) (Rates, error) {
	fid := slog.String("fid", "console.configs.GetRates")

	collection := firestore.Client.Collection("configs")
	if collection == nil {
		logCtx.Warn("configs collection not found", fid)
		return Rates{}, common.ErrNotFound{}
	}

	var (
		doc *fs.DocumentSnapshot
		err error
	)

	doc, err = collection.Doc("rates").Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		if !errors.Is(err, common.ErrNotFound{}) {
			logCtx.Error("unable to get bank rates config", fid, "error", err)
			return Rates{}, err
		}
	}

	r := Rates{}

	if err := doc.DataTo(&r); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read configs rates data", fid, "error", err)
		return Rates{}, err
	}

	return r, nil
}
