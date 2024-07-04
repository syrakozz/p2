package completion

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"disruptive/lib/common"
	"disruptive/lib/openai"
)

func chatOpenAI(ctx context.Context, logCtx *slog.Logger, config *Config, sessionMemory *SessionMemory, contextPrompt, query string, verbose bool) error {
	req := openai.ChatRequest{
		Model:      config.Model,
		Messages:   []openai.ChatMessage{{Role: "system", Content: contextPrompt}},
		Creativity: config.Creativity,
		MaxTokens:  250,
	}

	start := len(sessionMemory.Entries) - config.SessionMemory
	if start < 0 {
		start = 0
	}

	for _, e := range sessionMemory.Entries[start:] {
		req.Messages = append(req.Messages, openai.ChatMessage{Role: "user", Content: e.User})
		req.Messages = append(req.Messages, openai.ChatMessage{Role: "assistant", Content: e.Assistant})
	}

	req.Messages = append(req.Messages, openai.ChatMessage{
		Role:    "user",
		Content: strings.TrimSpace(fmt.Sprintf("%s. %s", query, config.QueryAttributes)),
	})

	if verbose {
		fmt.Println(common.MarshalIndent(req))
	}

	fmt.Printf("[%s: %s] %s\n\n", config.Set, config.Model, query)

	if config.Repeat > 0 {
		repeatCSV.Write([]string{fmt.Sprintf("[%s: %s] %s", config.Set, config.Model, query)})

		for i := 0; i < config.Repeat; i++ {
			t := time.Now()
			res, err := openai.PostChat(ctx, logCtx, req)
			if err != nil {
				logCtx.Error("unable to post chat", "error", err)
				return err
			}
			d := time.Since(t).Seconds()

			if verbose {
				fmt.Println(common.MarshalIndent(res))
			} else {
				fmt.Println("   ", res.Text)
				fmt.Println()
			}
			repeatCSV.Write([]string{"", res.Text, config.Set, config.Model, time.Now().Format("2006-01-02 15:04:05 MST"), fmt.Sprintf("%f", d), res.FinishReason, strconv.Itoa(res.Usage), strconv.Itoa(res.UsagePrompt), strconv.Itoa(res.UsageResponse)})
		}

	} else {
		res, err := openai.PostChat(ctx, logCtx, req)
		if err != nil {
			logCtx.Error("unable to post chat", "error", err)
			return err
		}

		if verbose {
			common.P(res)
		} else {
			fmt.Println(res.Text)
			fmt.Println()
		}

		sessionMemory.Entries = append(sessionMemory.Entries, SessionMemoryEntry{
			User:      query,
			Assistant: res.Text,
		})
	}

	return nil
}
