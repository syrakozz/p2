package play

import (
	"encoding/json"
	"log/slog"
	"os"

	"disruptive/lib/coqui"
	"disruptive/lib/elevenlabs"
)

type config struct {
	ElevenlabsVoices     map[string]elevenlabs.Voice `json:"11lab_voices"`
	CoquiVoices          map[string]string           `json:"coqui_voices"`
	CoquiXTTSVoices      map[string]string           `json:"coqui_xtts_voices"`
	Characters           map[string]character        `json:"characters"`
	SystemPromptTemplate string                      `json:"system_prompt_template"`
}

func readConfig(logCtx *slog.Logger) error {
	c := &config{}

	if b, err := os.ReadFile("vox.json"); err == nil {
		if err := json.Unmarshal(b, c); err != nil {
			logCtx.Error("unable unmarshal config.json", "error", err)
			return err
		}
	}

	for k, v := range c.ElevenlabsVoices {
		elevenlabs.Voices[k] = v
	}

	for k, v := range c.CoquiVoices {
		coqui.Voices[k] = v
	}

	for k, v := range c.CoquiXTTSVoices {
		coqui.VoicesXTTS[k] = v
	}

	for k, v := range c.Characters {
		characters[k] = v
	}

	if c.SystemPromptTemplate != "" {
		systemPromptTemplate = c.SystemPromptTemplate
	}

	return nil
}
