// Package elevenlabs integrates with the Elevenlabs API
package elevenlabs

import (
	"net/http"
	"slices"
	"time"

	"github.com/go-resty/resty/v2"

	"disruptive/config"
	"disruptive/lib/common"
)

const (
	baseURL                       = "https://api.elevenlabs.io/v1"
	voiceElevenLabsEndpoint       = "/text-to-speech/{voice}"
	voiceElevenLabsStreamEndpoint = "/text-to-speech/{voice}/stream"
)

type (
	// Request contains input parameters for TTS
	Request struct {
		Filename                 string
		Format                   string
		Language                 string
		Model                    string
		OptimizeStreamingLatency string
		SimilarityBoost          int
		Stability                int
		StyleExaggeration        int
		Text                     string
		Voice                    string
	}

	request struct {
		ModelID       string        `json:"model_id"`
		Text          string        `json:"text"`
		VoiceSettings voiceSettings `json:"voice_settings"`
	}

	voiceSettings struct {
		SimilarityBoost   float64 `json:"similarity_boost"`
		Stability         float64 `json:"stability"`
		StyleExaggeration float64 `json:"style"`
	}

	// Voice contains a voice from a config file
	Voice struct {
		ID                string `json:"id"`
		SimilarityBoost   int    `json:"similarity_boost"`
		Stability         int    `json:"stability"`
		StyleExaggeration int    `json:"style_exaggeration"`
	}
)

var (
	// AudioFormatExtensions is a map containing valid audio formats and their file extention.
	AudioFormatExtensions = map[string]string{
		"mp3_22050_32":  "mp3",
		"mp3_44100_32":  "mp3",
		"mp3_44100_64":  "mp3",
		"mp3_44100_96":  "mp3",
		"mp3_44100_128": "mp3",
		"mp3_44100_192": "mp3",
		"opus_16000":    "opus",
		"pcm_16000":     "l16",
		"pcm_22050":     "l22",
		"pcm_24000":     "l24",
		"pcm_44100":     "l44",
	}

	// AudioFormatContentTypes is a map containing valid audio formats and their contype types.
	AudioFormatContentTypes = map[string]string{
		"mp3_22050_32":  "audio/mpeg",
		"mp3_44100_32":  "audio/mpeg",
		"mp3_44100_64":  "audio/mpeg",
		"mp3_44100_96":  "audio/mpeg",
		"mp3_44100_128": "audio/mpeg",
		"mp3_44100_192": "audio/mpeg",
		"opus_16000":    "audio/opus",
		"pcm_16000":     "audio/pcm",
		"pcm_22050":     "audio/pcm",
		"pcm_24000":     "audio/pcm",
		"pcm_44100":     "audio/pcm",
	}

	// OptimizingStreamLatency contains possible option values.
	OptimizingStreamLatency = []string{"", "0", "1", "2", "3", "4"}

	languagesV1 = []string{"", "en", "en-AU", "en-GB", "en-NZ", "en-US", "ar", "de", "es", "es-419", "es-ES", "fr", "fr-CA", "fr-FR", "hi", "it", "pl", "pt", "pt-BR", "pt-PT"}
	languagesV2 = []string{"", "en", "en-AU", "en-GB", "en-NZ", "en-US", "ar", "bg", "cz", "de", "el", "es", "es-419", "es-ES", "fi", "fr", "fr-CA", "fr-FR", "hi", "hr", "id", "it", "ja", "ko", "ms", "nl", "pl", "pt", "pt-BR", "pt-PT", "ro", "sk", "sv", "ta", "tl", "tr", "uk", "zh"}

	// Resty is the shared Resty client for the elevenlabs package.
	Resty *resty.Client

	// Voices map names to voice IDs.
	Voices = map[string]Voice{}
)

func renderFromRequest(req Request) request {
	modelID := "eleven_multilingual_v2"
	if req.Model == "v1" && slices.Contains(languagesV1, req.Language) {
		modelID = "eleven_multilingual_v1"
	}

	return request{
		Text:    req.Text,
		ModelID: modelID,
		VoiceSettings: voiceSettings{
			SimilarityBoost:   float64(req.SimilarityBoost) / 100.0,
			Stability:         float64(req.Stability) / 100.0,
			StyleExaggeration: float64(req.StyleExaggeration) / 100.0,
		},
	}
}

func init() {
	Resty = resty.New().
		SetBaseURL(baseURL).
		SetLogger(common.LogDiscard).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		}).
		SetHeader("Xi-API-Key", config.VARS.ElevenLabsKey).
		SetHeader("User-Agent", config.VARS.UserAgent).
		SetHeader("Content-Type", "application/json").
		SetTimeout(time.Minute)

	Resty.GetClient().Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
		MaxConnsPerHost:     100,
	}
}
