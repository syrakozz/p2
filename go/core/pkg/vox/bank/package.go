// Package bank interfaces with the firestore user's bank collection.
package bank

import (
	"disruptive/lib/common"
	"disruptive/lib/firestore"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	baseURL             = "https://androidpublisher.googleapis.com/androidpublisher/v3"
	iapPurchaseEndpoint = "/applications/{packageName}/purchases/subscriptions/{subscriptionId}/tokens/{token}"
)

var (
	// Resty is the shared Resty client for the bank package.
	Resty *resty.Client
)

func init() {
	logCtx := slog.With("fid", "bank.init")

	if firestore.Client == nil {
		logCtx.Warn("unable use bank")
		return
	}

	Resty = resty.New().
		SetBaseURL(baseURL).
		SetLogger(common.LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		}).
		SetHeader("Content-Type", "application/json").
		SetTimeout(time.Minute)

	Resty.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
	}
}
