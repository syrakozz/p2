// Package highlevel integrates with HighLevel.
package highlevel

import (
	"context"
	"disruptive/config"
	"disruptive/lib/common"
	"disruptive/lib/firestore"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	baseURL              = "https://rest.gohighlevel.com/v1"
	contactsEndpoint     = "/contacts/"
	customFieldsEndpoint = "/custom-fields/"
)

var (
	// Resty is the shared Resty client for the highlevel package.
	Resty *resty.Client
)

func init() {
	logCtx := slog.With("fid", "highlevel.init")

	if firestore.Client == nil {
		logCtx.Warn("unable use highlevel")
		return
	}

	if err := LoadConfigs(context.Background(), logCtx); err != nil {
		logCtx.Warn("unable to load configs", "error", err)
	}

	Resty = resty.New().
		SetBaseURL(baseURL).
		SetLogger(common.LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		}).
		SetHeader("User-Agent", config.VARS.UserAgent).
		SetHeader("Content-Type", "application/json").
		SetTimeout(time.Minute)

	Resty.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
	}
}
