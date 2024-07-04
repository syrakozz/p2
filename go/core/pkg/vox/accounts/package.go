// Package accounts interfaces with the firestore user document.
package accounts

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"

	"disruptive/config"
	"disruptive/lib/common"
)

const (
	baseURL          = "https://api.mailgun.net/v3"
	messagesEndpoint = "/{domain}/messages"
)

var (
	// Resty is the shared resty client for the accounts package.
	Resty *resty.Client
)

func init() {
	Resty = resty.New().
		SetBaseURL(baseURL).
		SetLogger(common.LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests || r.StatusCode() == http.StatusInternalServerError
		}).
		SetBasicAuth("api", config.VARS.MailgunKey).
		SetHeader("User-Agent", config.VARS.UserAgent).
		SetHeader("Content-Type", "application/json").
		SetTimeout(time.Minute)

	Resty.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
	}
}
