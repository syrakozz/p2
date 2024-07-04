package completion

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
)

// Config contains the completions.json data.
type Config struct {
	Model           string   `json:"model"`
	Memory          string   `json:"memory"`
	SessionMemory   int      `json:"session_memory"`
	Failsafe        string   `json:"failsafe"`
	TopK            int      `json:"topk"`
	Creativity      int      `json:"creativity"`
	Moderation      bool     `json:"moderation"`
	Prompt          string   `json:"prompt"`
	Query           string   `json:"query"`
	QueryAttributes string   `json:"query_attributes"`
	Repeat          int      `json:"repeat"`
	RepeatFile      string   `json:"repeat_file"`
	Queries         []string `json:"queries"`
	Verbose         bool     `json:"verbose"`
	Set             string   `json:"set"`
	Sets            map[string]struct {
		Model           *string `json:"model"`
		Memory          *string `json:"memory"`
		SessionMemory   *int    `json:"session_memory"`
		Failsafe        *string `json:"failsafe"`
		TopK            *int    `json:"topk"`
		Creativity      *int    `json:"creativity"`
		Moderation      *bool   `json:"moderation"`
		Prompt          *string `json:"prompt"`
		Query           *string `json:"query"`
		QueryAttributes *string `json:"query_attributes"`
	} `json:"sets"`
}

func readConfig(logCtx *slog.Logger, configFile, set string) (*Config, error) {
	b, err := os.ReadFile(configFile)
	if err != nil {
		logCtx.Error("unable to read completions sets file", "error", err)
		return nil, err
	}

	config := &Config{}

	if err := json.Unmarshal(b, &config); err != nil {
		logCtx.Error("unable to unmarshal completions config file", "error", err)
		return nil, err
	}

	configSet, ok := config.Sets[set]
	if !ok {
		logCtx.Error("invalid completion set")
		return nil, errors.New("invalid completion set")
	}

	config.Set = set

	if configSet.Model != nil {
		config.Model = *configSet.Model
	}

	if configSet.Memory != nil {
		config.Memory = *configSet.Memory
	}

	if configSet.SessionMemory != nil {
		config.SessionMemory = *configSet.SessionMemory
	}

	if configSet.Failsafe != nil {
		config.Failsafe = *configSet.Failsafe
	}

	if configSet.TopK != nil {
		config.TopK = *configSet.TopK
	}

	if configSet.Creativity != nil {
		config.Creativity = *configSet.Creativity
	}

	if configSet.Moderation != nil {
		config.Moderation = *configSet.Moderation
	}

	if configSet.Prompt != nil {
		config.Prompt = *configSet.Prompt
	}

	if configSet.Query != nil {
		config.Query = *configSet.Query
	}

	if configSet.QueryAttributes != nil {
		config.QueryAttributes = *configSet.QueryAttributes
	}

	return config, nil
}
