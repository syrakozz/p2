// Package rocketreach integrates with RocketReach API
package rocketreach

import (
	"disruptive/config"
	"disruptive/lib/common"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	baseURL        = "https://api.rocketreach.co/api/v2"
	aboutEndpoint  = "/account/"
	lookupEndpoint = "/profile-company/lookup"
	searchEndpoint = "/search"
)

// Resty is the shared Resty client for the rocketreach package.
var Resty *resty.Client

func init() {
	Resty = resty.New().
		SetBaseURL(baseURL).
		SetLogger(common.LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		}).
		SetHeader("Api-Key", config.VARS.RocketReachKey).
		SetHeader("User-Agent", config.VARS.UserAgent).
		SetHeader("Content-Type", "application/json").
		SetTimeout(time.Minute)

	Resty.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
	}
}
