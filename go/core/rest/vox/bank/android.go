package bank

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"disruptive/lib/common"
	"disruptive/lib/configs"
	"disruptive/pkg/vox/accounts"
	"disruptive/pkg/vox/bank"
	"disruptive/rest/auth"
	e "disruptive/rest/errors"
)

// AndroidNotification contains the Android store notification.
type AndroidNotification struct {
	Message struct {
		Data      string `json:"data"` // Base64
		MessageID string `json:"messageId"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

// AndroidNotificationData contains the AndroidNotification data contents.
type AndroidNotificationData struct {
	Version                  string `json:"version"`
	PackageName              string `json:"packageName"`
	Time                     string `json:"eventTimeMillis"`
	SubscriptionNotification *struct {
		Version          string `json:"version"`
		SubscriptionType int    `json:"notificationType"`
		PurchaseToken    string `json:"purchaseToken"`
		SubcriptionID    string `json:"subscriptionId"`
	} `json:"subscriptionNotification"`
	OneTimeProductNotification *struct {
		Version          string `json:"version"`
		SubscriptionType int    `json:"notificationType"`
		PurchaseToken    string `json:"purchaseToken"`
		SKU              string `json:"sku"`
	} `json:"oneTimeProductNotification"`
	TestNotification *struct {
		Version string `json:"version"`
	} `json:"testNotification"`
}

// PostAndroidIAPTransaction is the REST API for creating Android IAP transactions.
func PostAndroidIAPTransaction(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.bank.PostAndroidIAPTransaction")
	account := ctx.Value(common.AccountKey).(accounts.Document)

	req := bank.GoogleAndroidIAPTransaction{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read android iap transaction")
	}

	if req.PackageName == "" {
		return e.ErrBad(logCtx, fid, "missing packageName")
	}

	if req.PurchaseToken == "" {
		return e.ErrBad(logCtx, fid, "missing purchaseToken")
	}

	if req.ProductID == "" {
		return e.ErrBad(logCtx, fid, "missing productId (SKU)")
	}

	if req.Quantity == 0 {
		return e.ErrBad(logCtx, fid, "missing quantity")
	}

	txn, err := bank.PostAndroidIAPTransaction(ctx, logCtx, account.ID, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to create android iap transaction")
	}

	return c.JSON(http.StatusCreated, txn)
}

// PostAndroidSubTransaction is the REST API for creating Android subscription transactions.
func PostAndroidSubTransaction(c echo.Context) error {
	ctx, logCtx, fid := auth.InitRequest(c, "rest.vox.bank.PostAndroidSubTransaction")
	account := ctx.Value(common.AccountKey).(accounts.Document)

	req := bank.AndroidSubTransaction{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read android transaction")
	}

	req.AccountID = account.ID
	req.Timestamp = time.Now()

	countries, err := configs.Get(ctx, logCtx, "countries")
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to get countries config file")
	}

	req.CountryCode = strings.ToUpper(req.CountryCode)
	country, ok := countries[req.CountryCode]
	if !ok {
		return e.ErrBad(logCtx, fid, "invalid country code")
	}

	req.LoadBalancer = country.(map[string]any)["load_balancer"].(string)

	if req.PurchaseToken == "" {
		return e.ErrBad(logCtx, fid, "missing purchase_token")
	}

	if req.SKU == "" {
		return e.ErrBad(logCtx, fid, "missing SKU")
	}

	if req.SubscriptionID == "" {
		return e.ErrBad(logCtx, fid, "missing subscription_id")
	}

	txn, err := bank.PostAndroidSubTransaction(ctx, logCtx, req)
	if err != nil {
		return e.Err(logCtx, err, fid, "unable to create android sub transaction")
	}

	return c.JSON(http.StatusCreated, txn)
}

// PostAndroidNotifications is the REST API for Android store notifications.
func PostAndroidNotifications(c echo.Context) error {
	fid := slog.String("fid", "rest.vox.bank.PostAndroidNotifications")
	logCtx := slog.With("sid", c.Response().Header().Get(echo.HeaderXRequestID))

	req := AndroidNotification{}

	if err := c.Bind(&req); err != nil {
		return e.ErrBad(logCtx, fid, "unable to read android notification")
	}

	if req.Message.Data == "" {
		return e.ErrBad(logCtx, fid, "invalid message data")
	}

	dataDecoded, err := base64.StdEncoding.DecodeString(req.Message.Data)
	if err != nil {
		return e.ErrBad(logCtx, fid, "unable to read android notification data")
	}

	data := AndroidNotificationData{}
	if err := json.Unmarshal(dataDecoded, &data); err != nil {
		return e.ErrBad(logCtx, fid, "unable to unmarshal data")
	}

	if data.TestNotification != nil {
		logCtx.Info("android store",
			"purchase", "test",
			"version", data.TestNotification.Version,
		)
		return c.NoContent(http.StatusOK)
	}

	if data.SubscriptionNotification != nil {
		logCtx.Info("android store",
			"purchase", "subscription",
			"version", data.SubscriptionNotification.Version,
			"type", data.SubscriptionNotification.SubscriptionType,
			"token", data.SubscriptionNotification.PurchaseToken,
			"sku", data.SubscriptionNotification.SubcriptionID,
		)
	} else if data.OneTimeProductNotification != nil {
		logCtx.Info("android store",
			"purchase", "one_time_product",
			"version", data.OneTimeProductNotification.Version,
			"type", data.OneTimeProductNotification.SubscriptionType,
			"token", data.OneTimeProductNotification.PurchaseToken,
			"sku", data.OneTimeProductNotification.SKU,
		)
	} else {
		return e.ErrBad(logCtx, fid, "invalid data")
	}

	return c.NoContent(http.StatusOK)
}
