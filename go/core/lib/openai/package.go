// Package openai integrates with openai.com
package openai

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"

	"disruptive/config"
	"disruptive/lib/common"
)

const (
	baseURL                = "https://api.openai.com/v1"
	chatEndpoint           = "/chat/completions"
	embeddingsEndpoint     = "/embeddings"
	fileEndpoint           = "/files/{file_id}"
	filesEndpoint          = "/files"
	finetunesEndpoint      = "/fine-tunes"
	finetuneEndpoint       = "/fine-tunes/{finetune_id}"
	modelsEndpoint         = "/models"
	modelEndpoint          = "/models/{model_id}"
	moderationsEndpoint    = "/moderations"
	transcriptionsEndpoint = "/audio/transcriptions"
	translationsEndpoint   = "/audio/translations"
)

var (
	// Rate limiters.
	textAda002Limiter = rate.NewLimiter(50, 50)
	whisper1Limiter   = rate.NewLimiter(rate.Every(300*time.Millisecond), 5)

	// Resty is the shared resty client for the openai package.
	Resty *resty.Client

	// AudioFormatExtensions is a map containing valid audio formats and their file extention.
	AudioFormatExtensions = map[string]string{
		"flac": "flac",
		"mp3":  "mp3",
		"m4a":  "m4a",
		"mp4":  "mp4",
		"mpeg": "mpeg",
		"mpga": "mpga",
		"ogg":  "ogg",
		"wav":  "wav",
		"webm": "webm",
	}
)

type errorResponse struct {
	Error struct {
		Message string  `json:"message"`
		Type    string  `json:"type"`
		Param   string  `json:"param"`
		Code    *string `json:"code"`
	} `json:"error"`
}

func init() {
	Resty = resty.New().
		SetBaseURL(baseURL).
		SetLogger(common.LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests ||
				r.StatusCode() == http.StatusInternalServerError ||
				r.StatusCode() == http.StatusServiceUnavailable
		}).
		SetAuthToken(config.VARS.OpenAIKey).
		SetHeader("User-Agent", config.VARS.UserAgent).
		SetHeader("Content-Type", "application/json").
		SetTimeout(time.Minute)

	Resty.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
	}
}

func errorMessage(res []byte) string {
	e := errorResponse{}
	if err := json.Unmarshal(res, &e); err != nil {
		return "error"
	}
	return e.Error.Message
}
