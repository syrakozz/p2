package common

import "fmt"

var (
	// AudioFormatExtensions is a map containing valid formats and their file extention.
	AudioFormatExtensions = map[string]string{
		"mp3":           "mp3",
		"mp3_44100":     "mp3",
		"mp3_44100_64":  "mp3",
		"mp3_44100_96":  "mp3",
		"mp3_44100_128": "mp3",
		"mp3_44100_192": "mp3",
		"opus":          "opus",
		"pcm_16000":     "l16",
		"pcm_22050":     "l22",
		"pcm_24000":     "l24",
		"pcm_44100":     "l44",
	}

	// AudioExtensionContentTypes is a map containing valid formats and their content types.
	AudioExtensionContentTypes = map[string]string{
		"flac": "audio/flac",
		"mp3":  "audio/mpeg",
		"m4a":  "audio/mp4",
		"mp4":  "audio/mp4",
		"mpeg": "audio/mpeg",
		"mpga": "audio/mpeg",
		"ogg":  "audio/ogg",
		"opus": "audio/opus",
		"wav":  "audio/wav",
		"webm": "audio/webm",
	}
)

// IntToMACAddress converts a uint64 into a valid MAC address
func IntToMACAddress(i uint64) string {
	// Ensure the integer is within the 48-bit range
	if i > 0xFFFFFFFFFFFF {
		return "Error: Integer out of MAC address range"
	}

	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		byte(i>>40),
		byte(i>>32),
		byte(i>>24),
		byte(i>>16),
		byte(i>>8),
		byte(i),
	)
}
