// Package stabilityai integrates with the Stablility AI APIs.
package stabilityai

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"

	"disruptive/config"
	"disruptive/lib/common"
)

const (
	baseURL             = "https://api.stability.ai/v1/"
	textToImageEndpoint = "generation/{engine_id}/text-to-image"
)

var (
	// Resty is the shared Resty client for the StableDiffusion package.
	Resty *resty.Client

	// DownloadResty is the shared Resty client for downloading the image.
	DownloadResty *resty.Client

	// Engines are the available engine sizes.
	Engines = map[string]string{
		"512":  "stable-diffusion-v1-6",
		"1024": "stable-diffusion-xl-1024-v1-0",
	}
)

func init() {
	Resty = resty.New().
		SetBaseURL(baseURL).
		SetAuthToken(config.VARS.StabilityAIKey).
		SetLogger(common.LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		}).
		SetTimeout(30 * time.Second)

	Resty.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
	}

	DownloadResty = resty.New().
		SetLogger(common.LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		}).
		SetTimeout(30 * time.Second)

	DownloadResty.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
	}
}
