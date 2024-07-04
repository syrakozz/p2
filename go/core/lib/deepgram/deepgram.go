package deepgram

import (
	"context"
	"disruptive/config"
	"io"
	"log/slog"
	"slices"
	"time"

	prerecorded "github.com/deepgram/deepgram-go-sdk/pkg/api/prerecorded/v1"
	interfaces "github.com/deepgram/deepgram-go-sdk/pkg/client/interfaces"
	client "github.com/deepgram/deepgram-go-sdk/pkg/client/prerecorded"
)

var (
	cli *client.Client

	// Languages contains the supported languages
	Languages = []string{
		"en", "en-AU", "en-GB", "en-IN", "en-NZ", "en-US", "da", "de", "es", "es-419", "fr", "fr-CA", "id",
		"it", "hi", "ja", "ko", "nl", "no", "pl",
		"pt", "pt-BR", "ru", "sv", "ta", "tr", "uk", "zh",
	}

	// LanguageCountry maps country codes to the default.
	LanguageCountry = map[string]string{
		"fr-FR": "fr",
		"pt-PT": "pt",
		"es-ES": "es",
	}

	// LanguagesWhisper contains languages that will be forced to use Whisper
	LanguagesWhisper = []string{
		"ar", "bg", "cz", "el", "fi", "hr",
		"ms", "ro", "sk", "th", "tl",
	}
)

func init() {
	client.Init(client.InitLib{
		LogLevel: client.LogLevelErrorOnly,
	})

	opt := interfaces.ClientOptions{ApiKey: config.VARS.DeepgramKey}
	if config.VARS.DeepgramHost != "" {
		opt.ApiKey = config.VARS.DeepgramOnPremKey
		opt.Host = config.VARS.DeepgramHost
		opt.OnPrem = true
	}

	cli = client.New("", opt)
	if cli == nil {
		slog.Warn("unable to create deepgram client")
		return
	}
}

// PostTranscriptionsText sends transcriptions to OpenAI and returns the text.
func PostTranscriptionsText(ctx context.Context, logCtx *slog.Logger, r io.Reader, language, version string) (string, string, error) {
	fid := slog.String("fid", "deepgram.PostTranscriptionText")

	if l := LanguageCountry[language]; l != "" {
		language = l
	}

	forceWhisper := false
	if slices.Contains(LanguagesWhisper, language) {
		forceWhisper = true
	}

	var options interfaces.PreRecordedTranscriptionOptions
	d := time.Second * 10

	if forceWhisper || language == "" {
		if config.VARS.DeepgramHost == "" {
			d = time.Second * 20
		}

		options = interfaces.PreRecordedTranscriptionOptions{
			Model:          "whisper-medium",
			DetectLanguage: true,
			Punctuate:      true,
		}
	} else {
		model := languageToModelV2[language]
		if model == "nova-2" && version == "v1" {
			model = languageToModelV1[language]
		}

		options = interfaces.PreRecordedTranscriptionOptions{
			Model:     model,
			Language:  language,
			Punctuate: true,
		}
	}

	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()

	res, err := prerecorded.New(cli).FromStream(ctx, r, options)
	if err != nil {
		logCtx.Error("unable to transcribe", fid, "error", err)
		return "", "", convertDeepgramError(err)
	}

	if len(res.Results.Channels) < 1 || len(res.Results.Channels[0].Alternatives) < 1 || res.Results.Channels[0].Alternatives[0].Transcript == "" {
		return "", "", nil
	}

	text := res.Results.Channels[0].Alternatives[0].Transcript
	if text == "" {
		logCtx.Warn("unable to retrieve transcript", fid)
	}

	detectedLanguage := ""
	if forceWhisper || language == "" {
		detectedLanguage = res.Results.Channels[0].DetectedLanguage
		if dl, ok := languages[detectedLanguage]; ok {
			detectedLanguage = dl
		}
	}

	return text, detectedLanguage, nil
}
