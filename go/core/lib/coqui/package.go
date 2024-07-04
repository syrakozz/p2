// Package coqui integrates with the Coqui API
package coqui

import (
	"disruptive/config"
	"disruptive/lib/common"
	"errors"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	baseURL                  = "https://app.coqui.ai/api/v2"
	sampleFileEndpoint       = "/samples"
	sampleFileXTTSEndpoint   = "/samples/xtts/render/"
	samplePromptXTTSEndpoint = "/samples/xtts/render-from-prompt/"

	maxSampleLength     = 500
	maxSampleXTTSLength = 250
)

type (
	// Request contains input parameters for TTS
	Request struct {
		Prompt string
		Voice  string
		Text   string
	}

	renderRequest struct {
		Voice  string `json:"voice_id,omitempty"`
		Prompt string `json:"prompt,omitempty"`
		Name   string `json:"name,omitempty"`
		Text   string `json:"text"`
		Speed  int    `json:"speed"`
	}

	renderResponse struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		CreatedAt time.Time `json:"created_at"`
		Text      string    `json:"text"`
		AudioURL  *string   `json:"audio_url"`
	}
)

var (
	// Resty is the shared Resty client for the coqui package.
	Resty *resty.Client

	// Voices map names to voice IDs.
	Voices = map[string]string{
		"johnny5": "41d34181-3601-4e39-a955-37a09c9d960c",
	}

	// VoicesXTTS map names to XXTS voice IDs.
	VoicesXTTS = map[string]string{
		"abrahan mack":      "b1ec84ad-c7c6-4085-b3e9-fcae55529b77",
		"alison dietlinde":  "de21e9e3-2da0-478e-a3d8-4b042d3a3b28",
		"ana florence":      "f05c5b91-7540-4b26-b534-e820d43065d1",
		"andrew chipper":    "81bad083-72dc-4071-8548-70ba944b8039",
		"annmarie nele":     "ff34248d-1fae-479b-85b6-9ae2b6043acd",
		"asya anara":        "e34ac3b4-0aed-4a7f-adf5-f2a2e2424694",
		"badr odhiambo":     "b67f6eb1-c3ac-4da2-b359-47895eb93580",
		"baldur sanjin":     "e399c204-7040-4f1d-bb92-5223fa6feceb",
		"brenda stern":      "9b1cb1b4-f4fa-48ea-af20-54c91f35bfdd",
		"clarabel dervla":   "b8ffb895-79b8-4ec6-be9c-6eb2d1fbe83c",
		"craig gutsy":       "27373d4a-0b84-480d-9ce3-fc34fba415be",
		"daisy studious":    "90ae3b42-29b1-479c-b012-846b0b640c72",
		"damian black1":     "62223f22-bd98-441f-b254-44925ff449c3",
		"damian black2":     "4847fb4f-e0bc-417c-98d7-1da389a79d7c",
		"damian black3":     "04262b81-6321-4d22-bbcf-1800bad30d94",
		"damian black4":     "0f82817b-eea7-4f28-8a02-5900a1b23e30",
		"damian black5":     "c390f2dd-ea36-49ff-ad55-179ab3a3b4ee",
		"damian black6":     "7d926c53-cb08-4fc7-9f05-021d426b4f69",
		"damian black7":     "e895d2c5-e77d-4189-a336-14c9f9d95035",
		"dionisio schuyler": "6720d486-5d43-4d92-8893-57a1b58b334d",
		"gilberto mathias":  "ba43f07b-67bf-47a2-bce5-b1d5fa2ba1b5",
		"gitta nikolina":    "d91d2f95-1a1d-4062-bad1-f1497bb5b487",
		"gracie wise":       "6ec4d93b-1f54-4420-91f8-33f188ee61f3",
		"henriette usha":    "8255e841-3b5c-48af-9089-640a2ee2c308",
		"ilkin urbano":      "8ca72d29-f9ec-4df8-8ad0-de7a1c5790b0",
		"kazuhiko atallah":  "ab86648c-68d3-4b03-a6dc-f4a78cf527d5",
		"ludvig milivoj":    "e1a51d31-0f2f-4532-98d4-7b73e2481d06",
		"royston min":       "67c19643-429d-4cef-bb30-bf2a84ba1c84",
		"sofia hellen":      "ebe2db86-62a6-49a1-907a-9a1360d4416e",
		"suad qasim":        "b082061d-695e-4d1b-a8f9-b5c4cb8e6e2a",
		"tammie ema":        "fd613b67-e9b8-45ae-8702-a34ff65f1b78",
		"tammy grit":        "9145d03c-2da9-4893-8c92-ee9480e75830",
		"tanja adelina":     "f6d81c82-1376-4dd5-9825-cd9f353cbfb9",
		"torcull diarmuid":  "d4b43fc7-6e16-4664-b9ec-97246f505d8d",
		"viktor eka":        "c791b5b5-0558-42b8-bb0b-602ac5efc0b9",
		"viktor menelaos":   "d2bd7ccb-1b65-4005-9578-32c4e02d8ddf",
		"vjollca johnnie":   "cb4f835e-7f61-4b8c-a0f6-f059bbf6f583",
		"zacharie aimilios": "fc9917ef-8f32-418e-9254-e535c0c6df3d",
	}
)

func init() {
	Resty = resty.New().
		SetBaseURL(baseURL).
		SetLogger(common.LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		}).
		SetAuthToken(config.VARS.CoquiKey).
		SetHeader("User-Agent", config.VARS.UserAgent).
		SetHeader("Content-Type", "application/json").
		SetTimeout(time.Minute)

	Resty.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
	}
}

func getVoice(voice string) (string, string, error) {
	v, ok := Voices[voice]
	if ok {
		return v, sampleFileEndpoint, nil
	}

	v, ok = VoicesXTTS[voice]
	if ok {
		return v, sampleFileXTTSEndpoint, nil
	}

	return "", "", errors.New("invalid voice")
}
