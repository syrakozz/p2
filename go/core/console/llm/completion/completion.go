// Package completion tests the completion API for OpenAI and Anthropic.
package completion

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"disruptive/lib/pinecone"
)

type (
	// SessionMemoryEntry is a single session memory entry pair.
	SessionMemoryEntry struct {
		User      string `json:"user"`
		Assistant string `json:"assistant"`
	}

	// SessionMemory contains multiple session memory entry pairs.
	SessionMemory struct {
		Entries []SessionMemoryEntry `json:"entries"`
	}
)

var (
	repeatCSV *csv.Writer
)

// Main integrates Pinecone context, Session memory and feeds it to OpenAI/Anthropic Chat.
func Main(ctx context.Context, configFile, set, query string, repeat int, repeatFile string, verbose bool) error {
	logCtx := slog.With("config_file", configFile, "set", set, "query", query)

	config, err := readConfig(logCtx, configFile, set)
	if err != nil {
		logCtx.Error("unable to read config", "error", err)
		return err
	}

	config.Verbose = verbose

	if query != "" {
		config.Query = query
	}

	if len(query) > 5 && strings.HasPrefix(query, "file:") {
		if err := readQueryFile(config); err != nil {
			logCtx.Error("unable to read query file", "error", err, "file", config.Query[5:])
			return err
		}
	} else {
		config.Queries = append(config.Queries, config.Query)
	}

	if repeat > 0 && repeatFile == "" {
		logCtx.Error("invalid repeat file")
		return errors.New("invalid repeat file")
	}

	if repeat > 0 {
		config.Repeat = repeat
	}

	if repeatFile != "" {
		config.RepeatFile = repeatFile
	}

	if config.Repeat > 0 {
		exists := false
		if _, err := os.Stat(config.RepeatFile); err == nil {
			exists = true
		}

		f, err := os.OpenFile(config.RepeatFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logCtx.Error("unable to create repeat file", "error", err, "file", config.RepeatFile)
			return err
		}
		defer f.Close()

		repeatCSV = csv.NewWriter(f)
		defer repeatCSV.Flush()

		if !exists {
			repeatCSV.Write([]string{"Question", "Answer", "Set", "Model", "Time", "Duration (seconds)", "Finish Reason", "Tokens", "Prompt Tokens", "Response Tokens"})
			repeatCSV.Flush()
		}
	}

	logCtx = logCtx.With(
		"model", config.Model,
		"memory", config.Memory,
		"topk", config.TopK,
		"creativity", config.Creativity,
		"moderation", config.Moderation,
	)

	// Session Memory

	sessionMemory := &SessionMemory{}

	if b, err := os.ReadFile(set + ".json"); err == nil {
		if err := json.Unmarshal(b, sessionMemory); err != nil {
			logCtx.Warn("unable to unmarshal session memory", "error", err)
			return err
		}
	}

	for _, q := range config.Queries {
		if config.Moderation {
			if err := processModeration(ctx, logCtx, q); err != nil {
				return err
			}
		}

		if err := processQuery(ctx, logCtx, config, sessionMemory, q); err != nil {
			logCtx.Error("unable to process query", "error", err, "query", q)
			return err
		}
	}

	b, err := json.MarshalIndent(sessionMemory, "", "  ")
	if err != nil {
		logCtx.Error("unable to marshal session memory", "error", err)
		return err
	}

	if err := os.WriteFile(set+".json", b, 0644); err != nil {
		logCtx.Error("unable to write session memory", "error", err)
		return err
	}

	return nil
}

func readQueryFile(config *Config) error {
	f, err := os.Open(config.Query[5:])
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		config.Queries = append(config.Queries, strings.TrimSpace(scanner.Text()))
	}

	return scanner.Err()
}

func processQuery(ctx context.Context, logCtx *slog.Logger, config *Config, sessionMemory *SessionMemory, query string) error {
	var (
		queryRes []pinecone.QueryResponse
		err      error
	)

	if config.Memory != "" {
		queryRes, err = processMemory(ctx, logCtx, config.Memory, query, config.TopK, config.Verbose)
		if err != nil {
			logCtx.Error("unable to get memory", "error", err)
			return errors.New("unable to get memory")
		}
	}

	// Memory Context
	var contextPrompt strings.Builder
	contextPrompt.WriteString(config.Prompt)
	contextPrompt.WriteString(" ")
	contextPrompt.WriteString(config.Failsafe)

	if config.Memory != "" {
		contextPrompt.WriteString("\n\nCONTEXT:\n")

		for _, v := range queryRes {
			q := v.Metadata["question"]
			a := v.Metadata["answer"]
			contextPrompt.WriteString(fmt.Sprintf("Question: %s\nAnswer: %s\n\n", q, a))
		}
	}

	switch config.Model {
	case "gpt-3.5-turbo", "gpt-4":
		chatOpenAI(ctx, logCtx, config, sessionMemory, contextPrompt.String(), query, config.Verbose)
	case "claude-v1", "claude-instant-v1":
		chatAnthropic(ctx, logCtx, config, sessionMemory, contextPrompt.String(), query, config.Verbose)
	default:
		logCtx.Error("unknown model")
	}

	return nil
}
