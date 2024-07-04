// Package accounts interfaces with the firestore user document.
package accounts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	fs "cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"disruptive/lib/common"
	"disruptive/lib/configs"
	"disruptive/lib/firebase"
	"disruptive/lib/firestore"
)

// Document contains a firestore account document.
type Document struct {
	ID               string             `firestore:"id" json:"id"` // account ID is the same as the account's firebase_id
	Admin            bool               `firestore:"admin" json:"admin"`
	CreatedDate      time.Time          `firestore:"created_date" json:"created_date"`
	DeveloperMode    bool               `firestore:"developer_mode" json:"developer_mode"`
	DeveloperModeMap map[string]any     `firestore:"developer_mode_map" json:"developer_mode_map"`
	DisableBank      bool               `firestore:"disable_bank" json:"disable_bank,omitempty"`
	DisplayName      string             `firestore:"display_name" json:"display_name"`
	Email            string             `firestore:"email" json:"email"`
	Inactive         bool               `firestore:"inactive" json:"inactive"`
	ModifiedDate     time.Time          `firestore:"modified_date" json:"modified_date"`
	Preferences      map[string]any     `firestore:"preferences" json:"preferences"`
	Products         map[string]Product `firestore:"products" json:"products"`
	Pin              string             `firestore:"pin" json:"pin,omitempty"`
	Timezone         string             `firestore:"timezone" json:"timezone"`
}

// PatchDocument contains a firestore account patch document.
type PatchDocument struct {
	ID               string          `firestore:"id" json:"id"`
	DeveloperModeMap *map[string]any `firestore:"developer_mode_map" json:"developer_mode_map"`
	DisplayName      *string         `firestore:"display_name" json:"display_name"`
	Email            *string         `firestore:"email" json:"email"`
	Inactive         *bool           `firestore:"inactive" json:"inactive"`
	Pin              *string         `firestore:"pin" json:"pin"`
	Timezone         *string         `firestore:"timezone" json:"timezone"`
}

// Product contains account product attributes.
type Product struct {
	IDs []string `firestore:"ids" json:"ids"`
}

// ProductDeviceConnect contains information for connecting a product device
type ProductDeviceConnect struct {
	Balance               int  `json:"balance"`
	FirstTime             bool `json:"first_time"`
	FirstTimeBalanceAdded int  `json:"first_time_balance_added,omitempty"`
	ValidID               bool `json:"valid_id"`
}

// GetAccount retrieves a firestore account.
func GetAccount(ctx context.Context, logCtx *slog.Logger, uid string) (Document, error) {
	fid := slog.String("fid", "vox.accounts.GetAccount")

	d := Document{}

	collection := firestore.Client.Collection("accounts")
	if collection == nil {
		logCtx.Error("unable to get account collection", fid)
		return d, common.ErrNotFound{}
	}

	doc, err := collection.Doc(uid).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get accounts document", fid, "error", err)
		return d, err
	}

	if err := doc.DataTo(&d); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read account data", fid, "error", err)
		return d, err
	}

	return d, nil
}

// GetAccountByEmail retrieves a firestore account by email.
func GetAccountByEmail(ctx context.Context, logCtx *slog.Logger, email string) (Document, error) {
	fid := slog.String("fid", "vox.accounts.GetAccountByEmail")

	collection := firestore.Client.Collection("accounts")
	if collection == nil {
		logCtx.Error("unable to get account collection", fid)
		return Document{}, common.ErrNotFound{}
	}

	iter := collection.Where("email", "==", email).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			err = common.ConvertGRPCError(err)
			return Document{}, err
		}

		d := Document{}
		if err := doc.DataTo(&d); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to read profile data", fid, "error", err)
			return Document{}, err
		}

		return d, nil
	}

	return Document{}, common.ErrNotFound{Msg: "account not found"}
}

// CreateAccount creates a new firestore account document.
func CreateAccount(ctx context.Context, logCtx *slog.Logger, document Document) (Document, error) {
	fid := slog.String("fid", "vox.accounts.CreateAccount")

	collection := firestore.Client.Collection("accounts")
	if collection == nil {
		logCtx.Error("unable to get account collection", fid)
		return Document{}, common.ErrNotFound{}
	}

	document.CreatedDate = time.Now()

	if _, err := collection.Doc(document.ID).Create(ctx, document); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to create accounts document", fid, "error", err)
		return Document{}, err
	}

	if err := CreateBalanceDocument(ctx, logCtx, document.ID); err != nil {
		logCtx.Warn("unable to get create account balance document", fid, "error", err)
		return Document{}, err
	}

	a, err := GetAccount(ctx, logCtx, document.ID)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get account document", fid, "error", err)
		return Document{}, err
	}

	return a, nil
}

// PatchAccount updates a firestore account document.
func PatchAccount(ctx context.Context, logCtx *slog.Logger, document PatchDocument) (Document, error) {
	fid := slog.String("fid", "vox.accounts.PatchAccount")

	collection := firestore.Client.Collection("accounts")
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return Document{}, common.ErrNotFound{}
	}

	updates := []fs.Update{}

	if document.DeveloperModeMap != nil {
		for k, v := range *document.DeveloperModeMap {
			if v == nil {
				v = fs.Delete
			}

			updates = append(updates, fs.Update{Path: "developer_mode_map." + k, Value: v})
		}
	}

	if document.DisplayName != nil {
		updates = append(updates, fs.Update{Path: "display_name", Value: *document.DisplayName})
	}

	if document.Email != nil {
		updates = append(updates, fs.Update{Path: "email", Value: *document.Email})
	}

	if document.Inactive != nil {
		updates = append(updates, fs.Update{Path: "inactive", Value: *document.Inactive})
	}

	if document.Pin != nil {
		updates = append(updates, fs.Update{Path: "pin", Value: *document.Pin})
	}

	if document.Timezone != nil {
		updates = append(updates, fs.Update{Path: "timezone", Value: *document.Timezone})
	}

	if len(updates) > 0 {
		updates = append(updates, fs.Update{Path: "modified_date", Value: time.Now()})

		if _, err := collection.Doc(document.ID).Update(ctx, updates); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Warn("unable to update account document", fid, "error", err)
			return Document{}, err
		}
	}

	a, err := GetAccount(ctx, logCtx, document.ID)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to get account document", fid, "error", err)
		return Document{}, err
	}

	return a, nil
}

// DeleteAccount removes an account from firestore.
func DeleteAccount(ctx context.Context, _ *slog.Logger, uid string) error {
	return firebase.DeleteUser(ctx, uid)
}

// PutProducts a firestore account document products.
func PutProducts(ctx context.Context, logCtx *slog.Logger, uid, product string, value Product) error {
	fid := slog.String("fid", "vox.accounts.PutProducts")

	collection := firestore.Client.Collection("accounts")
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return common.ErrNotFound{}
	}

	updates := []fs.Update{
		{Path: fmt.Sprintf("products.%s", product), Value: value},
	}

	if _, err := collection.Doc(uid).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to update account document product", fid, "error", err)
		return err
	}

	return nil
}

// ConnectProductDevice verifies and connects a product device to an account.
func ConnectProductDevice(ctx context.Context, logCtx *slog.Logger, product, deviceID string) (ProductDeviceConnect, error) {
	fid := slog.String("fid", "vox.accounts.ConnectProductDevice")

	account := ctx.Value(common.AccountKey).(Document)

	var (
		bal       BalanceDocument
		firstTime bool
		validID   bool
	)

	// check if the product deviceID already exists on the account.
	if slices.Contains(account.Products[product].IDs, deviceID) {
		bal, err := GetBalance(ctx, logCtx)
		if err != nil {
			logCtx.Error("unable to get account balance document", fid)
			return ProductDeviceConnect{}, common.ErrNotFound{}
		}

		return ProductDeviceConnect{
			Balance:   bal.Balance,
			FirstTime: false,
			ValidID:   true,
		}, nil
	}

	// product deviceID does not exist on the account.

	cfg, err := configs.Get(ctx, logCtx, "products")
	if err != nil {
		logCtx.Error("unable to get character config", fid)
		return ProductDeviceConnect{}, common.ErrNotFound{}
	}

	productCfg, ok := cfg[product].(map[string]any)
	if !ok {
		logCtx.Error("invalid product config", fid, "error", err)
		return ProductDeviceConnect{}, err
	}

	// check if device is an iOS firmware 2.0 UUID. if so, handle first time vexels and return.
	// do not add UUIDs to account or product list.
	if strings.Contains(deviceID, "-") {
		pdc, err := firstTimeIOSFirmware20(ctx, logCtx, productCfg)
		if err != nil {
			logCtx.Error("unable to verify and connect IOS Firmware 2.0 UUID device", fid)
			return ProductDeviceConnect{}, err
		}

		return pdc, nil
	}

	// deviceID is a MAC address. verify the MAC address.

	productIDPath := fmt.Sprintf("products/%s/ids", product)
	collection := firestore.Client.Collection(productIDPath)
	if collection == nil {
		logCtx.Error("configs character IDs collection not found", fid)
		return ProductDeviceConnect{}, common.ErrNotFound{}
	}

	productDeviceDoc, err := collection.Doc(deviceID).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		if errors.Is(err, common.ErrNotFound{}) {
			// deviceID does not exist on factory list

			whiteList, ok := productCfg["white_list"].(bool)
			if !ok || !whiteList {
				logCtx.Error("invalid characterID", fid, "error", err)
				return ProductDeviceConnect{}, err
			}

			// whitelist=true, add deviceID to factory ID list.
			if _, err := collection.Doc(deviceID).Set(ctx, map[string]any{"account_id": account.ID, "timestamp": time.Now(), "white_list": true}); err != nil {
				err = common.ConvertGRPCError(err)
				logCtx.Error("unable to add character_id to factory list", fid, "error", err)
				return ProductDeviceConnect{}, err
			}
			validID = true
			firstTime = true
		} else {
			logCtx.Error("unable to get character config document", fid, "error", err)
			return ProductDeviceConnect{}, err
		}
	}

	if productDeviceDoc.Exists() {
		// device ID already exists on the factory ID list

		validID = true

		deviceDoc := map[string]any{}
		if err := productDeviceDoc.DataTo(&deviceDoc); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to read product config data", fid, "error", err)
			return ProductDeviceConnect{}, err
		}

		// check if product device is already paired to an account.
		if _, ok := deviceDoc["account_id"].(string); !ok {
			firstTime = true
		}
	}

	var productDoc ProductDeviceConnect
	if validID {
		// add valid deviceID to account product ID list
		collection = firestore.Client.Collection("accounts")
		if collection == nil {
			logCtx.Error("unable to get accounts collection", fid)
			return ProductDeviceConnect{}, common.ErrNotFound{}
		}

		updates := []fs.Update{
			{Path: fmt.Sprintf("products.%s.ids", product), Value: fs.ArrayUnion(deviceID)},
		}

		if _, err := collection.Doc(account.ID).Update(ctx, updates); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to update account document character", fid, "error", err)
			return ProductDeviceConnect{}, err
		}

		if firstTime {
			// add timestamp to factory deviceID and the accountID of the account it is assigned to.
			updates = []fs.Update{
				{Path: "account_id", Value: account.ID},
				{Path: "timestamp", Value: time.Now()},
			}

			collection = firestore.Client.Collection(productIDPath)
			if _, err := collection.Doc(deviceID).Update(ctx, updates); err != nil {
				err = common.ConvertGRPCError(err)
				logCtx.Error("unable to add deviceID to factory list", fid, "error", err)
				return ProductDeviceConnect{}, err
			}

			// add any first time promo codes to account
			if promoCodes, ok := productCfg["first_time_promo_codes"].([]any); ok {
				for _, code := range promoCodes {
					if c, ok := code.(string); ok {
						res, err := RedeemGiftCard(ctx, logCtx, c)
						if err != nil {
							logCtx.Warn("unable to get add promo code to account", fid, "gift_card", c)
							continue
						}

						productDoc.FirstTimeBalanceAdded += res.ValueRedeemed
					}
				}
			}
		}
	}

	bal, err = GetBalance(ctx, logCtx)
	if err != nil {
		logCtx.Error("unable to get account balance document", fid)
		return ProductDeviceConnect{}, common.ErrNotFound{}
	}

	productDoc.Balance = bal.Balance
	productDoc.FirstTime = firstTime
	productDoc.ValidID = validID

	return productDoc, nil
}

// DeleteCharacter a firestore account document character.
func DeleteCharacter(ctx context.Context, logCtx *slog.Logger, uid, characterName string) error {
	fid := slog.String("fid", "vox.accounts.DeleteCharacter")

	collection := firestore.Client.Collection("accounts")
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return common.ErrNotFound{}
	}

	updates := []fs.Update{
		{Path: fmt.Sprintf("characters.%s", characterName), Value: fs.Delete},
	}

	if _, err := collection.Doc(uid).Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to delete account document character", fid, "error", err)
		return err
	}

	return nil
}

// handle IOS firmware 2.0 connection.
// treat UUID as white-listed, update/check balance's "first_time_ios_firmware_20" value
func firstTimeIOSFirmware20(ctx context.Context, logCtx *slog.Logger, productCfg map[string]any) (ProductDeviceConnect, error) {
	fid := slog.String("fid", "vox.accounts.firstTimeIOSFirmware20")

	bal, err := GetBalance(ctx, logCtx)
	if err != nil {
		logCtx.Error("unable to get account balance document", fid)
		return ProductDeviceConnect{}, common.ErrNotFound{}
	}

	// check if account has already receieved its first time balance.
	if bal.FirstTimeIOSFirmware20 {
		return ProductDeviceConnect{
			Balance:   bal.Balance,
			FirstTime: false,
			ValidID:   true,
		}, nil
	}

	// account has not received its first time balance.
	// add any first time promo codes to account
	productDoc := ProductDeviceConnect{
		Balance:   bal.Balance,
		FirstTime: true,
		ValidID:   true,
	}

	if promoCodes, ok := productCfg["first_time_promo_codes"].([]any); ok {
		for _, code := range promoCodes {
			if c, ok := code.(string); ok {
				res, err := RedeemGiftCard(ctx, logCtx, c)
				if err != nil {
					logCtx.Warn("unable to get add promo code to account", fid, "gift_card", c)
					continue
				}

				productDoc.Balance += res.ValueRedeemed
				productDoc.FirstTimeBalanceAdded += res.ValueRedeemed
			}
		}
	}

	account := ctx.Value(common.AccountKey).(Document)

	path := fmt.Sprintf("accounts/%s/bank", account.ID)
	collection := firestore.Client.Collection(path)
	if collection == nil {
		logCtx.Error("unable to get accounts collection", fid)
		return ProductDeviceConnect{}, common.ErrNotFound{}
	}

	updates := []fs.Update{{Path: "first_time_ios_firmware_20", Value: true}}

	if _, err := collection.Doc("balance").Update(ctx, updates); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Warn("unable to update accounts bank document", fid, "error", err)
		return ProductDeviceConnect{}, err
	}

	return productDoc, nil
}
