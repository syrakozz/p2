// Package deepgram integrations with the Deepgram APIs.
package deepgram

var (
	// AudioFormatExtensions is a map containing valid audio formats and their file extention.
	AudioFormatExtensions = map[string]string{
		"flac": "flac",
		"mp3":  "mp3",
		"m4a":  "m4a",
		"mp4":  "mp4",
		"mpeg": "mpeg",
		"mpga": "mpga",
		"opus": "opus",
		"ogg":  "ogg",
		"wav":  "wav",
		"webm": "webm",
	}

	languages = map[string]string{
		"":   "en-US",
		"en": "en-US",
		"es": "es-ES",
		"fr": "fr-FR",
		"pt": "pt-PT",
	}

	// nova > enhanced > base > whisper
	languageToModelV1 = map[string]string{
		"en":     "nova",
		"en-AU":  "nova",
		"en-GB":  "nova",
		"en-IN":  "nova",
		"en-NZ":  "nova",
		"en-US":  "nova",
		"de":     "enhanced",
		"es":     "nova",
		"es-419": "nova",
		"es-ES":  "nova",
		"fr":     "enhanced",
		"fr-CA":  "enhanced",
		"fr-FR":  "enhanced",
		"hi":     "enhanced",
		"pt":     "enhanced",
		"pt-BR":  "enhanced",
		"pt-PT":  "enhanced",
	}

	// nova > enhanced > base > whisper
	languageToModelV2 = map[string]string{
		"en":     "nova-2",
		"en-AU":  "nova-2",
		"en-GB":  "nova-2",
		"en-IN":  "nova-2",
		"en-NZ":  "nova-2",
		"en-US":  "nova-2",
		"da":     "nova-2",
		"de":     "nova-2",
		"es":     "nova-2",
		"es-419": "nova-2",
		"es-ES":  "nova-2",
		"fr":     "nova-2",
		"fr-CA":  "nova-2",
		"fr-FR":  "nova-2",
		"id":     "nova-2",
		"it":     "nova-2",
		"hi":     "nova-2",
		"ja":     "enhanced",
		"ko":     "nova-2",
		"nl":     "nova-2",
		"no":     "nova-2",
		"pl":     "nova-2",
		"pt":     "nova-2",
		"pt-BR":  "nova-2",
		"pt-PT":  "nova-2",
		"ru":     "nova-2",
		"sv":     "nova-2",
		"ta":     "enhanced",
		"tr":     "nova-2",
		"uk":     "nova-2",
		"zh":     "base",
	}
)
