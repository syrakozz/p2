// Package mailfinder integrates with the anymailfinder service
package mailfinder

import (
	"disruptive/config"
	"disruptive/lib/common"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	baseURL                              = "https://api.anymailfinder.com"
	anymailfinderAccountEndpoint         = "/v5.0/meta/account.json"
	anymailfinderSearchEmployeesEndpoint = "/v5.0/search/employees.json"
	anymailfinderStatusEndpoint          = "/status"
)

var (
	// Resty is the shared Resty client for the mailfinder package.
	Resty *resty.Client
)

func init() {
	Resty = resty.New().
		SetBaseURL(baseURL).
		SetLogger(common.LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		}).
		SetAuthToken(config.VARS.AnyMailFinderKey).
		SetHeader("User-Agent", config.VARS.UserAgent).
		SetHeader("Content-Type", "application/json").
		SetTimeout(time.Minute)

	Resty.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
	}
}
