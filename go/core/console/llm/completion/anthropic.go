package completion

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"disruptive/lib/anthropic"
	"disruptive/lib/common"
)

func chatAnthropic(ctx context.Context, logCtx *slog.Logger, config *Config, sessionMemory *SessionMemory, contextPrompt, query string, verbose bool) error {
	var buf strings.Builder

	buf.WriteString(contextPrompt)

	start := len(sessionMemory.Entries) - config.SessionMemory
	if start < 0 {
		start = 0
	}

	for _, e := range sessionMemory.Entries[start:] {
		buf.WriteString(fmt.Sprintf("Human: %s\n\nAssistant:%s\n\n", e.User, e.Assistant))
	}

	buf.WriteString(fmt.Sprintf("Human: %s. %s\n\nAssistant:", query, config.QueryAttributes))

	req := anthropic.CompletionRequest{
		Model:       config.Model,
		Prompt:      buf.String(),
		MaxTokens:   250,
		Temperature: float64(config.Creativity) / 100.0,
	}

	if verbose {
		fmt.Println(common.MarshalIndent(req))
	}

	fmt.Printf("[%s: %s] %s\n\n", config.Set, config.Model, query)

	if config.Repeat > 0 {
		repeatCSV.Write([]string{fmt.Sprintf("[%s: %s] %s", config.Set, config.Model, query)})

		for i := 0; i < config.Repeat; i++ {
			t := time.Now()
			res, err := anthropic.Completion(ctx, logCtx, req)
			if err != nil {
				logCtx.Error("unable to execute completion", "error", err)
				return err
			}
			d := time.Since(t).Seconds()

			res.Completion = strings.TrimSpace(res.Completion)

			if verbose {
				fmt.Println(common.MarshalIndent(res))
			} else {
				fmt.Println("   ", res.Completion)
				fmt.Println()
			}

			if res.StopReason == "stop_sequence" {
				res.StopReason = "stop"
			}
			repeatCSV.Write([]string{"", res.Completion, config.Set, config.Model, time.Now().Format("2006-01-02 15:04:05 MST"), fmt.Sprintf("%f", d), res.StopReason})
		}

	} else {
		res, err := anthropic.Completion(ctx, logCtx, req)
		if err != nil {
			logCtx.Error("unable to execute completion", "error", err)
			return err
		}

		res.Completion = strings.TrimSpace(res.Completion)

		if verbose {
			fmt.Println(common.MarshalIndent(res))
		} else {
			fmt.Println("   ", res.Completion)
			fmt.Println()
		}

		sessionMemory.Entries = append(sessionMemory.Entries, SessionMemoryEntry{
			User:      query,
			Assistant: res.Completion,
		})

	}

	return nil
}
