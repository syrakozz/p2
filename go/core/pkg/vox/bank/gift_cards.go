package bank

import (
	"context"
	"log/slog"
	"time"

	fs "cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"disruptive/lib/common"
	"disruptive/lib/firestore"
)

// GiftCards are a map of gift cards
type GiftCards map[string]GiftCardInfo

// GiftCardInfo contains a promo codes information
type GiftCardInfo struct {
	Description string    `firestore:"description" json:"description"`
	Expiration  time.Time `firestore:"expiration" json:"expiration"`
	OneTimeUse  bool      `firestore:"one_time_use" json:"one_time_use"`
	Start       time.Time `firestore:"start" json:"start"`
	Value       int64     `firestore:"value" json:"value"`
}

// GetGiftCard retrieves a current gift card.
func GetGiftCard(ctx context.Context, logCtx *slog.Logger, giftCard string) (GiftCardInfo, error) {
	fid := slog.String("fid", "vox.bank.GetGiftCards")

	var collection *fs.CollectionRef

	collection = firestore.Client.Collection("bank/gift_cards/current")
	if collection == nil {
		logCtx.Error("unable to get current gift cards collection", fid)
		return GiftCardInfo{}, common.ErrNotFound{}
	}

	doc, err := collection.Doc(giftCard).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get gift card document", fid, "error", err)
		return GiftCardInfo{}, err
	}

	g := GiftCardInfo{}

	if err := doc.DataTo(&g); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read gift card data", fid, "error", err)
		return GiftCardInfo{}, err
	}

	if g.Expiration.Before(time.Now()) && !time.Time.IsZero(g.Expiration) {
		if _, err := doc.Ref.Delete(ctx); err != nil {
			logCtx.Error("unable to delete gift card", fid, "error", err)
			return GiftCardInfo{}, err
		}

		if _, err := firestore.Client.Collection("bank/gift_cards/expired").Doc(giftCard).Set(ctx, g); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to add gift card to expired", fid, "error", err)
			return GiftCardInfo{}, err
		}
	}

	return g, nil
}

// GetGiftCards retrieves current or expired gift cards.
func GetGiftCards(ctx context.Context, logCtx *slog.Logger, expired string) (GiftCards, error) {
	fid := slog.String("fid", "vox.bank.GetGiftCards")

	var collection *fs.CollectionRef

	if expired == "true" {
		collection = firestore.Client.Collection("bank/gift_cards/expired")
		if collection == nil {
			logCtx.Error("unable to get expired gift cards collection", fid)
			return GiftCards{}, common.ErrNotFound{}
		}
	} else {
		collection = firestore.Client.Collection("bank/gift_cards/current")
		if collection == nil {
			logCtx.Error("unable to get current gift cards collection", fid)
			return GiftCards{}, common.ErrNotFound{}
		}
	}

	giftCards := GiftCards{}
	expiredCards := GiftCards{}

	iter := collection.Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logCtx.Error("unable to retrieve gift card documents", fid)
			return GiftCards{}, err
		}

		gc := GiftCardInfo{}
		if err := doc.DataTo(&gc); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to read gift card data", fid, "error", err)
			return GiftCards{}, err
		}

		if expired != "true" && gc.Expiration.Before(time.Now()) && !time.Time.IsZero(gc.Expiration) {
			expiredCards[doc.Ref.ID] = gc
			if _, err := doc.Ref.Delete(ctx); err != nil {
				logCtx.Error("unable to delete gift card", fid, "error", err)
				return GiftCards{}, err
			}
		} else {
			giftCards[doc.Ref.ID] = gc
		}
	}

	for name, data := range expiredCards {
		if _, err := firestore.Client.Collection("bank/gift_cards/expired").Doc(name).Set(ctx, data); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to add gift card to expired", fid, "error", err)
			return GiftCards{}, err
		}
	}

	return giftCards, nil
}

// ExpireGiftCard moves a current gift card to expired.
func ExpireGiftCard(ctx context.Context, logCtx *slog.Logger, giftCard string) error {
	fid := slog.String("fid", "vox.bank.ExpireGiftCard")

	collection := firestore.Client.Collection("bank/gift_cards/current")
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return common.ErrNotFound{}
	}

	doc, err := collection.Doc(giftCard).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get current gift card document", fid, "error", err)
		return err
	}

	gc := GiftCardInfo{}
	if err := doc.DataTo(&gc); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read gift card data", fid, "error", err)
		return err
	}

	if _, err := doc.Ref.Delete(ctx); err != nil {
		logCtx.Error("unable to delete gift card", fid, "error", err)
		return err
	}

	if _, err := firestore.Client.Collection("bank/gift_cards/expired").Doc(giftCard).Set(ctx, gc); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to add gift card to expired", fid, "error", err)
		return err
	}

	return nil
}
