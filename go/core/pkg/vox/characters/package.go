// Package characters interfaces with Firestore character collection.
package characters

import (
	"context"
	"log/slog"

	"disruptive/lib/configs"
)

// Character contains the values from a firestore character document.
type Character struct {
	Character      string            `firestore:"character,omitempty" json:"character,omitempty"`
	DontSay        []string          `firestore:"dont_say,omitempty" json:"dont_say,omitempty"`
	DontSayName    bool              `firestore:"dont_say_name,omitempty" json:"dont_say_name,omitempty"`
	Engine         string            `firestore:"engine,omitempty" json:"engine,omitempty"`
	LongName       string            `firestore:"long_name,omitempty" json:"long_name,omitempty"`
	Model          string            `firestore:"model,omitempty" json:"model,omitempty"`
	Modes          map[string]Mode   `firestore:"modes,omitempty" json:"modes,omitempty"`
	Name           string            `firestore:"name,omitempty" json:"name,omitempty"`
	TraitsNegative []string          `firestore:"traits_negative,omitempty" json:"traits_negative,omitempty"`
	TraitsPositive []string          `firestore:"traits_positive,omitempty" json:"traits_positive,omitempty"`
	Version        int               `firestore:"version,omitempty" json:"version,omitempty"`
	Voices         map[string]string `firestore:"voices,omitempty" json:"voices,omitempty"`
}

// Mode contains the character mode values.
type Mode struct {
	CharacterPrompt string `firestore:"character_prompt" json:"character_prompt,omitempty"`
	Creativity      int    `firestore:"creativity" json:"creativity,omitempty"`
	MaxWords        int    `firestore:"max_words" json:"max_words,omitempty"`
	SessionEntries  int    `firestore:"session_entries" json:"session_entries,omitempty"`
	Tier            string `firestore:"tier" json:"tier,omitempty"`
}

const (
	minSessionArchiveEntries = 200
	keepSessionEntries       = 15
	maxNumberEntries         = "999999"

	// MaxPromptTokens35Turbo is the maximum tokens we accept for gpt-3.5-turbo
	MaxPromptTokens35Turbo = 12288

	// MaxPromptTokensGPT4Turbo is the maximum tokens we accept for gpt-4-turbo 128k model
	MaxPromptTokensGPT4Turbo = 102400

	// GPT4Turbo is the current gpt-4-turbo model
	GPT4Turbo = "gpt-4-turbo-preview"

	// TokenEncodingModel is the tiktoken encoding model
	TokenEncodingModel = "cl100k_base"
)

var (
	characters = map[string]Character{}
)

func init() {
	if err := configs.SetElevenlabsVoices(context.Background(), slog.Default()); err != nil {
		slog.Warn("unable to get elevenlabs_voices config", slog.String("fid", "vox.init"), "error", err)
	}
}
